package delivery

import (
	"context"
	"dota_league/model"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type SourceService interface {
	Inspect(context.Context, string, int, string) (*model.SourceInspection, error)
	Coverage(context.Context, int, int, string) ([]model.SourceStatus, int64, error)
}

func NewSourceDelivery(e *echo.Echo, service SourceService) {
	e.GET("/source-data/coverage", func(c echo.Context) error {
		meta := newMeta(c)
		if meta.Limit == 0 {
			meta.Limit = 20
		}
		rows, total, err := service.Coverage(c.Request().Context(), meta.Offset, meta.Limit, c.QueryParam("search"))
		if err != nil {
			return echo.NewHTTPError(502, "Could not load collection coverage")
		}
		meta.Total = total
		return c.JSON(200, response{Meta: meta, Results: rows})
	})
	e.GET("/source-data/:kind/:id", func(c echo.Context) error {
		kind := c.Param("kind")
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 || (kind != "team" && kind != "league") {
			return echo.NewHTTPError(400, "Invalid source record")
		}
		result, err := service.Inspect(c.Request().Context(), kind, id, c.QueryParam("version"))
		if err != nil {
			return echo.NewHTTPError(502, "Could not load source data")
		}
		if c.QueryParam("download") == "1" {
			if result.Snapshot == nil {
				return echo.NewHTTPError(404, "Snapshot not available")
			}
			c.Response().Header().Set("Content-Disposition", "attachment; filename=\""+kind+"-"+strconv.Itoa(id)+".json\"")
			// JSON only, served as an attachment; never interpret source strings as HTML.
			return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, json.RawMessage(result.Snapshot.Payload))
		}
		return c.JSON(200, result)
	})
}
