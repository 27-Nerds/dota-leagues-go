package delivery

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

// NewStaticDelivery serves assets with the site logo as a fallback for leagues
// whose logo is not available locally. The fallback is never saved as a league
// logo, so the downloader can still retrieve the real image on a later refresh.
func NewStaticDelivery(e *echo.Echo, root string) {
	static := echo.StaticDirectoryHandler(os.DirFS(root), true)
	e.GET("/*", func(c echo.Context) error {
		path := c.Param("*")
		id, isLogo := strings.CutSuffix(path, "/logo.png")
		if isLogo && positiveID(id) {
			if _, err := os.Stat(filepath.Join(root, id, "logo.png")); os.IsNotExist(err) {
				c.Response().Header().Set("Cache-Control", "public, max-age=60")
				return c.File(filepath.Join(root, "logo.png"))
			} else if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
			}
		}
		return static(c)
	})
}
