package api

import (
	"context"
	"errors"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func TestInteractiveRequestDoesNotWaitBehindBackgroundBacklog(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		previous := valveLimiter.Load()
		defer valveLimiter.Store(previous)
		previousGate := backgroundRequests
		backgroundRequests = make(chan struct{}, 1)
		defer func() { backgroundRequests = previousGate }()
		limiter := rate.NewLimiter(1, 1)
		valveLimiter.Store(limiter)
		if !limiter.Allow() {
			t.Fatal("initial token unavailable")
		}
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		var workers sync.WaitGroup
		for range 30 {
			workers.Go(func() { _ = waitForValve(ctx, false) })
		}
		synctest.Wait()
		start := time.Now()
		if err := waitForValve(ctx, true); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed != 2*time.Second {
			t.Fatalf("interactive wait = %s; want 2s while respecting the shared 1 RPS limit", elapsed)
		}
		cancel()
		workers.Wait()
		if len(backgroundRequests) != 0 {
			t.Fatal("cancelled workers leaked the background gate")
		}
	})
}

// Steam lookups run on their own limiter, so a Valve backlog must not delay them.
func TestSteamRequestIgnoresValveBacklog(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		previous := valveLimiter.Load()
		defer valveLimiter.Store(previous)
		previousGate := backgroundRequests
		backgroundRequests = make(chan struct{}, 1)
		defer func() { backgroundRequests = previousGate }()
		valveLimiter.Store(rate.NewLimiter(rate.Every(time.Hour), 1))
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		var workers sync.WaitGroup
		for range 5 {
			workers.Go(func() { _ = waitForValve(ctx, false) })
		}
		synctest.Wait()
		start := time.Now()
		if err := steamLimiter.Load().Wait(ctx); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(start); elapsed != 0 {
			t.Fatalf("steam wait = %s; want no delay from the Valve backlog", elapsed)
		}
		cancel()
		workers.Wait()
	})
}

func TestRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := doRequest(ctx, "http://unused.invalid"); !errors.Is(err, context.Canceled) {
		t.Fatalf("rate-limit wait did not propagate cancellation: %v", err)
	}

	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel = context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		body, err := doRequest(ctx, server.URL)
		if body != nil {
			closeResponse(body)
		}
		done <- err
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("request did not reach the server")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("HTTP request did not propagate cancellation: %v", err)
	}
}

func TestIncompleteImageIsNotPublished(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		if _, err := io.WriteString(w, "partial"); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	dir := t.TempDir()
	if err := DownloadImageIfNotExist(t.Context(), server.URL, dir, "logo.png"); err == nil {
		t.Fatal("expected an incomplete download error")
	}
	if _, err := os.Stat(filepath.Join(dir, "logo.png")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("incomplete image was published: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatalf("temporary download was not cleaned up: %v / %v", entries, err)
	}
}
