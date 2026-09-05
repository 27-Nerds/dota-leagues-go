package delivery

import (
	e "dota_league/error"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

// LeaguesDelivery struct
type LeaguesDelivery struct {
	LeaguesHandler LeaguesService
	GamesHandler   GamesService
}

// NewLeaguesDelivery adds routes to echo
func NewLeaguesDelivery(e *echo.Echo, lh LeaguesService, gh GamesService) {
	leaguesDelivery := &LeaguesDelivery{
		LeaguesHandler: lh,
		GamesHandler:   gh,
	}

	e.GET("/leagues", leaguesDelivery.getAllActive)
	e.GET("/leagues/:id/live-games", leaguesDelivery.getLiveGames)
	e.GET("/leagues/:id/series", leaguesDelivery.getSeries)
	e.GET("/leagues/:id", leaguesDelivery.getByID)
}

func (ld *LeaguesDelivery) getSeries(c echo.Context) error {
	meta := newMeta(c)
	leagueID := c.Param("id")

	seriesFromDB, totalCount, err := ld.LeaguesHandler.GetSeries(c.Request().Context(), leagueID, meta.Offset, meta.Limit)
	if e.IsNotFound(err) {
		return c.JSON(http.StatusNotFound, "League series not found")
	} else if err != nil {
		slog.ErrorContext(c.Request().Context(), "get league series", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	meta.Total = totalCount

	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: seriesFromDB,
	})
}

func (ld *LeaguesDelivery) getAllActive(c echo.Context) error {
	meta := newMeta(c)
	leaguesFromDB, totalCount, err := ld.LeaguesHandler.GetAllActive(c.Request().Context(), meta.Offset, meta.Limit)
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "get active leagues", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}
	meta.Total = totalCount

	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: generateLeaguesDetailsResponse(leaguesFromDB),
	})
}

func (ld *LeaguesDelivery) getLiveGames(c echo.Context) error {
	meta := newMeta(c)
	id := c.Param("id")
	gamesFromDB, totalCount, err := ld.GamesHandler.GetLiveLeagueGames(c.Request().Context(), id, meta.Offset, meta.Limit)
	if e.IsNotFound(err) {
		slog.DebugContext(c.Request().Context(), "live games not found", "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "League Not Found Or No Live Games At the Moment")
	} else if err != nil {
		slog.ErrorContext(c.Request().Context(), "get live games", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}
	meta.Total = totalCount

	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: generateGameResponse(gamesFromDB),
	})
}

func (ld *LeaguesDelivery) getByID(c echo.Context) error {
	id := c.Param("id")
	league, err := ld.LeaguesHandler.GetByID(c.Request().Context(), id)

	if e.IsNotFound(err) {
		slog.DebugContext(c.Request().Context(), "league not found", "error", err)
		return echo.NewHTTPError(http.StatusNotFound, "League Not Found")
	} else if err != nil {
		slog.ErrorContext(c.Request().Context(), "get league", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, generateLeagueDetailsResponse(league))
}
