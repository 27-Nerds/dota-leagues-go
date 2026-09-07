package api

import (
	"context"
	e "dota_league/error"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// valveLimiter is shared by all Valve API requests.
var valveLimiter atomic.Pointer[rate.Limiter]

// steamLimiter paces steamcommunity.com requests. Steam is a separate host with
// its own limits, so its lookups neither wait for nor delay Valve API traffic.
var steamLimiter atomic.Pointer[rate.Limiter]

// Only one background request may reserve a future token. Other workers wait
// here so interactive cache misses do not queue behind every live-game poll.
var backgroundRequests = make(chan struct{}, 1)

func init() {
	valveLimiter.Store(rate.NewLimiter(defaultValveRPS, 1))
	steamLimiter.Store(rate.NewLimiter(defaultSteamRPS, 1))
}

const defaultValveRPS = 1.5
const defaultSteamRPS = 1.0

var httpClient = &http.Client{Timeout: 15 * time.Second}

// SetValveRateLimit configures the shared rate limiter for all Valve API requests (requests per second).
// Call it before starting workers.
func SetValveRateLimit(rps float64) {
	valveLimiter.Store(newLimiter(rps))
	slog.Info("Valve API rate limit configured", "rps", rps)
}

// SetSteamRateLimit configures the limiter for Steam Community requests (requests per second).
// Call it before starting workers.
func SetSteamRateLimit(rps float64) {
	steamLimiter.Store(newLimiter(rps))
	slog.Info("Steam rate limit configured", "rps", rps)
}

func newLimiter(rps float64) *rate.Limiter {
	burst := 1
	if b := int(rps); b > burst {
		burst = b
	}
	return rate.NewLimiter(rate.Limit(rps), burst)
}

func doRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	return doRateLimitedRequest(ctx, url, false)
}

// doSteamRequest bypasses the Valve limiter and background slot entirely.
func doSteamRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	if err := steamLimiter.Load().Wait(ctx); err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: "api.doSteamRequest", Err: err}
	}
	return performRequest(ctx, url)
}

func doInteractiveRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	return doRateLimitedRequest(ctx, url, true)
}

func waitForValve(ctx context.Context, interactive bool) error {
	if !interactive {
		select {
		case backgroundRequests <- struct{}{}:
			defer func() { <-backgroundRequests }()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return valveLimiter.Load().Wait(ctx)
}

func doRateLimitedRequest(ctx context.Context, url string, interactive bool) (io.ReadCloser, error) {
	op := "api.doRequest"

	if err := waitForValve(ctx, interactive); err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	return performRequest(ctx, url)
}

func performRequest(ctx context.Context, url string) (io.ReadCloser, error) {
	op := "api.doRequest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	req.Header.Set("User-Agent", "Valve/Steam HTTP Client 1.0 (570)")

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		closeResponse(res.Body)
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Message: fmt.Sprintf("upstream HTTP status %d", res.StatusCode)}
	}
	return res.Body, nil
}

// DownloadImageIfNotExist publishes an image only after its download completes.
func DownloadImageIfNotExist(ctx context.Context, sourceURL, relativeDestPath, imageName string) error {
	const op = "api.DownloadImageIfNotExist"
	dest := filepath.Join(relativeDestPath, imageName)
	if _, err := os.Stat(dest); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return &e.Error{Op: op, Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return &e.Error{Op: op, Err: err}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return &e.Error{Op: op, Err: err}
	}
	defer closeResponse(resp.Body)
	if resp.StatusCode != http.StatusOK {
		code := e.EINTERNAL
		if resp.StatusCode == http.StatusNotFound {
			code = e.ENOTFOUND
		}
		return &e.Error{Code: code, Op: op, Message: fmt.Sprintf("upstream HTTP status %d", resp.StatusCode)}
	}
	if err := os.MkdirAll(relativeDestPath, 0755); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	out, err := os.CreateTemp(relativeDestPath, ".image-*")
	if err != nil {
		return &e.Error{Op: op, Err: err}
	}
	defer func() {
		if err := os.Remove(out.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.WarnContext(ctx, "remove temporary image", "error", err)
		}
	}()
	_, copyErr := io.Copy(out, resp.Body)
	if err := errors.Join(copyErr, out.Close(), ctx.Err()); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	if err := os.Chmod(out.Name(), 0644); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	if err := os.Rename(out.Name(), dest); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	return nil
}

func closeResponse(body io.Closer) {
	if err := body.Close(); err != nil {
		slog.Warn("close API response", "error", err)
	}
}

// ValveRateLimit reports the configured request budget for background scheduling.
func ValveRateLimit() float64 { return float64(valveLimiter.Load().Limit()) }
