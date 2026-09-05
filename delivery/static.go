package delivery

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// NewStaticDelivery serves assets with the site logo as a fallback for leagues
// and teams whose logo is not available locally. The fallback is never saved as a
// logo, so the downloader can still retrieve the real image on a later refresh.
func NewStaticDelivery(e *echo.Echo, root string, assetDirectories ...string) {
	assetRoot := root
	if len(assetDirectories) > 0 && assetDirectories[0] != "" {
		assetRoot = assetDirectories[0]
	}
	static := echo.StaticDirectoryHandler(os.DirFS(root), true)
	serve := func(c echo.Context) error {
		path := c.Param("*")
		if c.Path() == "/teams/:id/logo.png" {
			if !positiveID(c.Param("id")) {
				return echo.ErrNotFound
			}
			path = "teams/" + c.Param("id") + "/logo.png"
		}
		id, isLogo := strings.CutSuffix(path, "/logo.png")
		id = strings.TrimPrefix(id, "teams/")
		if isLogo && positiveID(id) {
			if _, err := os.Stat(filepath.Join(assetRoot, path)); os.IsNotExist(err) {
				c.Response().Header().Set("Cache-Control", "public, max-age=60")
				return c.File(filepath.Join(root, "logo.png"))
			} else if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
			return c.File(filepath.Join(assetRoot, path))
		}
		if c.Path() == "/teams/:id/logo.png" {
			return c.File(filepath.Join(root, path))
		}
		return static(c)
	}
	e.GET("/*", serve)
	// An explicit route avoids the /teams/:id API route taking precedence.
	e.GET("/teams/:id/logo.png", serve)
}
