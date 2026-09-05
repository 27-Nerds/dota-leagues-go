// Package delivery exposes HTTP routes and formats API responses.
package delivery

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// DPCDelivery struct
type DPCDelivery struct {
	DPCHandler          StandingsService
	DPCResultsHandler   LeagueResultsService
	MatchMinimalHandler MatchService
}

// NewDPCDelivery adds routes to echo
func NewDPCDelivery(e *echo.Echo, dh StandingsService, rsh LeagueResultsService, mmh MatchService) {
	dpcDelivery := &DPCDelivery{
		DPCHandler:          dh,
		DPCResultsHandler:   rsh,
		MatchMinimalHandler: mmh,
	}

	e.GET("/dpc/standings", dpcDelivery.getStandings)
	e.GET("/leagues/:leagueID/results", dpcDelivery.getLeagueResults)
	e.GET("/leagues/:leagueID/matches/:matchID/minimal", dpcDelivery.getMatchMinimal)
}

func (dd *DPCDelivery) getStandings(c echo.Context) error {
	meta := newMeta(c)
	meta.Limit = 1
	meta.Offset = 0
	meta.Total = 1

	dpcFromDB, err := dd.DPCHandler.Get(c.Request().Context())
	if err != nil {
		slog.ErrorContext(c.Request().Context(), "get standings", "error", err)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: dpcFromDB,
	})
}

func (dd *DPCDelivery) getLeagueResults(c echo.Context) error {
	meta := newMeta(c)
	meta.Limit = 1
	meta.Offset = 0
	meta.Total = 1

	leagueID, err := strconv.Atoi(c.Param("leagueID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid league id")
	}

	resultsFromDB, herr := dd.DPCResultsHandler.Get(c.Request().Context(), leagueID)
	if herr != nil {
		slog.ErrorContext(c.Request().Context(), "get league results", "error", herr)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	meta.Total = int64(len(resultsFromDB.Results))
	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: resultsFromDB,
	})
}

func (dd *DPCDelivery) getMatchMinimal(c echo.Context) error {
	meta := newMeta(c)
	meta.Limit = 1
	meta.Offset = 0
	meta.Total = 1

	leagueID, err := strconv.Atoi(c.Param("leagueID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid league id")
	}

	matchID := c.Param("matchID")

	matchFromDB, herr := dd.MatchMinimalHandler.Get(c.Request().Context(), leagueID, matchID)
	if herr != nil {
		slog.ErrorContext(c.Request().Context(), "get match", "error", herr)
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, response{
		Meta:    meta,
		Results: matchFromDB,
	})
}
