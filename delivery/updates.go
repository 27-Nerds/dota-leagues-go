package delivery

import (
	"context"
	"dota_league/model"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type UpdatesService interface {
	GetAll(context.Context, int, int, ...model.UpdateFilter) ([]model.Update, int64, error)
}

type UpdatesDelivery struct{ service UpdatesService }

func NewUpdatesDelivery(e *echo.Echo, service UpdatesService) {
	d := &UpdatesDelivery{service: service}
	e.GET("/updates", d.getAll)
}

func (d *UpdatesDelivery) getAll(c echo.Context) error {
	meta := newMeta(c)
	if meta.Limit == 0 {
		meta.Limit = 10
	}
	filter := model.UpdateFilter{Search: strings.TrimSpace(c.QueryParam("search")), Entity: c.QueryParam("entity")}
	switch filter.Entity {
	case "", "roster", "team", "tournament", "player":
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid activity type")
	}
	if days := c.QueryParam("days"); days != "" {
		var err error
		filter.Days, err = strconv.Atoi(days)
		if err != nil || (filter.Days != 1 && filter.Days != 7 && filter.Days != 30) {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid time range")
		}
	}
	rows, total, err := d.service.GetAll(c.Request().Context(), meta.Offset, meta.Limit, filter)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "get updates", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}
	if rows == nil {
		rows = []model.Update{}
	}
	meta.Total = total
	return c.JSON(http.StatusOK, response{Meta: meta, Results: rows})
}
