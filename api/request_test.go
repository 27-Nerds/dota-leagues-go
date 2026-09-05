package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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
