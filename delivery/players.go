package delivery

import (
	e "dota_league/error"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type PlayersDelivery struct{ service PlayersService }

// NewPlayersDelivery adds player routes to echo
func NewPlayersDelivery(e *echo.Echo, service PlayersService) {
	d := &PlayersDelivery{service: service}
	e.GET("/players", d.getAll)
	e.GET("/players/:id", d.getOne)
}

func (d *PlayersDelivery) getAll(c echo.Context) error {
	meta := newMeta(c)
	filter, err := playerFilter(c)
	if err != nil {
		return err
	}
	rows, total, err := d.service.GetAll(c.Request().Context(), meta.Offset, meta.Limit, filter)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "get players", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}
	meta.Total = total
	return c.JSON(http.StatusOK, response{Meta: meta, Results: rows})
}

func (d *PlayersDelivery) getOne(c echo.Context) error {
	player, err := d.service.GetByID(c.Request().Context(), c.Param("id"))
	if e.IsNotFound(err) {
		return echo.NewHTTPError(http.StatusNotFound, "Player Not Found")
	} else if err != nil {
		slog.ErrorContext(c.Request().Context(), "get player", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}
	return c.JSON(http.StatusOK, player)
}
