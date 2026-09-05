package main

import (
	"dota_league/delivery"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestStaticFilesDoNotConsumeAPIQuota(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "logo.png"), []byte("fallback"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "20067"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "20067", "logo.png"), []byte("logo"), 0644); err != nil {
		t.Fatal(err)
	}

	e := echo.New()
	e.Use(IPRateLimitWithConfig(1, 0))
	delivery.NewStaticDelivery(e, root)
	e.GET("/leagues/:id", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.GET("/teams/:id", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	request := func(method, path string, want int) {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.RemoteAddr = "127.0.0.1:12345"
		e.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, rec.Code, want, rec.Body.String())
		}
	}

	for range 30 {
		request(http.MethodGet, "/20067/logo.png", http.StatusOK)
		request(http.MethodGet, "/teams/2163/logo.png", http.StatusOK)
	}
	request(http.MethodGet, "/leagues/20067", http.StatusOK)
	request(http.MethodGet, "/leagues/20067", http.StatusTooManyRequests)
	request(http.MethodGet, "/leagues/logo.png", http.StatusTooManyRequests)
	request(http.MethodGet, "/teams/2163", http.StatusTooManyRequests)
	request(http.MethodGet, "/teams/2163/logo.png", http.StatusOK)
	request(http.MethodGet, "/20067/logo.png", http.StatusOK)
	request(http.MethodGet, "/19979/logo.png", http.StatusOK)
	request(http.MethodGet, "/20067/missing.png", http.StatusNotFound)
}
