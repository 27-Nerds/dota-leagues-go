package delivery

import (
	e "dota_league/error"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TeamsDelivery struct
type TeamsDelivery struct {
	TeamsHandler TeamsService
}

// NewTeamsDelivery adds routes to echo
func NewTeamsDelivery(e *echo.Echo, th TeamsService) {
	teamsDelivery := &TeamsDelivery{
		TeamsHandler: th,
	}

	e.GET("/teams", teamsDelivery.getAll)
	e.GET("/teams/:id", teamsDelivery.getOne)
}

func (td *TeamsDelivery) getAll(c echo.Context) error {
	meta := newMeta(c)
	filter, err := teamFilter(c)
	if err != nil {
		return err
	}
	teamsFromDB, totalCount, err := td.TeamsHandler.GetAll(c.Request().Context(), meta.Offset, meta.Limit, filter)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "get teams", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}
	meta.Total = totalCount

	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: teamsFromDB,
	})
}

func (td *TeamsDelivery) getOne(c echo.Context) error {
	id := c.Param("id")
	team, err := td.TeamsHandler.GetByID(c.Request().Context(), id)
	if e.IsNotFound(err) {
		slog.DebugContext(c.Request().Context(), "team not found", "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "Team Not Found")
	} else if err != nil {
		slog.ErrorContext(c.Request().Context(), "get team", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, team)
}
