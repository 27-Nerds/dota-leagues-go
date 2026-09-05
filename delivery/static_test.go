package delivery

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestStaticLeagueLogoFallback(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "logo.png"), []byte("fallback"), 0644); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	NewStaticDelivery(e, root)
	check := func(path string, status int, body string) {
		t.Helper()
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != status || (status == http.StatusOK && rec.Body.String() != body) {
			t.Fatalf("GET %s: got %d %q, want %d %q", path, rec.Code, rec.Body.String(), status, body)
		}
	}
	check("/19979/logo.png", http.StatusOK, "fallback")
	check("/20067/logo.png", http.StatusOK, "fallback")
	check("/unknown/logo.png", http.StatusNotFound, "")
	check("/19979/missing.png", http.StatusNotFound, "")
	if _, err := os.Stat(filepath.Join(root, "19979", "logo.png")); !os.IsNotExist(err) {
		t.Fatalf("fallback must not create a cached league logo: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "19979"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "19979", "logo.png"), []byte("real logo"), 0644); err != nil {
		t.Fatal(err)
	}
	check("/19979/logo.png", http.StatusOK, "real logo")
}
