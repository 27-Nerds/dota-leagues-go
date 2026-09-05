package api

import (
	"context"
	e "dota_league/error"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// valveLimiter is shared by all Valve API requests; it holds a *rate.Limiter and is
// written once at startup (SetValveRateLimit) while workers read it concurrently.
var valveLimiter atomic.Value

func init() {
	SetValveRateLimit(defaultValveRPS)
}

const defaultValveRPS = 1.5

// SetValveRateLimit configures the shared rate limiter for all Valve API requests (requests per second).
// Call it before starting workers.
func SetValveRateLimit(rps float64) {
	burst := 1
	if b := int(rps); b > burst {
		burst = b
	}
	valveLimiter.Store(rate.NewLimiter(rate.Limit(rps), burst))
	log.Printf("valve api rate limit set to %.2f rps (burst %d)", rps, burst)
}

func doRequest(url string) (io.ReadCloser, error) {
	op := "api.doRequest"

	limiter := valveLimiter.Load().(*rate.Limiter)
	if err := limiter.Wait(context.Background()); err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	httpsClient := http.Client{
		Timeout: time.Second * 15, // Timeout after 15 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	req.Header.Set("User-Agent", "Valve/Steam HTTP Client 1.0 (570)")

	res, err := httpsClient.Do(req)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return res.Body, nil
}

// DownloadImageIfNotExist Download image
func DownloadImageIfNotExist(sourceURL string, relativeDestPath string, imageName string) error {
	op := "api.DownloadImageIfNotExist"
	basePath, err := os.Getwd()
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	path := fmt.Sprintf("%s/%s", basePath, relativeDestPath)

	fullPathWithName := fmt.Sprintf("%s/%s", path, imageName)

	_, err = os.Stat(fullPathWithName)
	if os.IsNotExist(err) {
		log.Printf("downloading image, %s", sourceURL)

		resp, err := http.Get(sourceURL)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

		defer closeResponse(resp.Body)

		if resp.StatusCode != 200 {
			return &e.Error{Code: e.ENOTFOUND, Op: op}
		}

		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}
		out, err := os.Create(fullPathWithName)
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}

		_, err = io.Copy(out, resp.Body)
		closeErr := out.Close()
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}
		if closeErr != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: closeErr}
		}

	}

	return nil
}

func closeResponse(body io.Closer) {
	if err := body.Close(); err != nil {
		log.Printf("close API response: %v", err)
	}
}
