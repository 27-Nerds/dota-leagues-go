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
	check("/teams/2163/logo.png", http.StatusOK, "fallback")
	check("/teams/unknown/logo.png", http.StatusNotFound, "")
	check("/teams/2163/missing.png", http.StatusNotFound, "")
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
	if err := os.MkdirAll(filepath.Join(root, "teams", "2163"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "teams", "2163", "logo.png"), []byte("team logo"), 0644); err != nil {
		t.Fatal(err)
	}
	check("/teams/2163/logo.png", http.StatusOK, "team logo")
}

func TestStaticDownloadsSeparateFromBuild(t *testing.T) {
	root, assets := t.TempDir(), t.TempDir()
	for path, body := range map[string]string{filepath.Join(root, "logo.png"): "fallback", filepath.Join(root, "bundle.js"): "new build", filepath.Join(assets, "bundle.js"): "stale download", filepath.Join(assets, "12", "logo.png"): "league", filepath.Join(assets, "teams", "34", "logo.png"): "team"} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	e := echo.New()
	NewStaticDelivery(e, root, assets)
	for path, want := range map[string]string{"/bundle.js": "new build", "/12/logo.png": "league", "/teams/34/logo.png": "team", "/99/logo.png": "fallback"} {
		r := httptest.NewRecorder()
		e.ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
		if r.Code != 200 || r.Body.String() != want {
			t.Fatalf("%s: %d %s", path, r.Code, r.Body)
		}
	}
}
