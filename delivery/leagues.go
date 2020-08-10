package delivery

import (
	e "dota_league/error"
	"dota_league/handler"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// LeaguesDelivery struct
type LeaguesDelivery struct {
	LeaguesHandler *handler.LeaguesHandlerInterface
	GamesHandler   *handler.GamesHandlerInterface
}

// NewLeaguesDelivery adds routes to echo
func NewLeaguesDelivery(e *echo.Echo, lh *handler.LeaguesHandlerInterface, gh *handler.GamesHandlerInterface) {
	leaguesDelivery := &LeaguesDelivery{
		LeaguesHandler: lh,
		GamesHandler:   gh,
	}

	e.GET("/leagues", leaguesDelivery.getAllActive)
	e.GET("/leagues/:id/live-games", leaguesDelivery.getLiveGames)
	e.GET("/leagues/:id", leaguesDelivery.get)
}

func (ld *LeaguesDelivery) getAllActive(c echo.Context) error {
	leagues, err := (*ld.LeaguesHandler).GetAllActive()
	if err != nil {
		log.Printf("getAllActive Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, leagues)
}

func (ld *LeaguesDelivery) getLiveGames(c echo.Context) error {
	id := c.Param("id")
	games, err := (*ld.GamesHandler).GetLiveLeagueGames(id)
	if e.IsNotFound(err) {
		log.Printf("getAllActive Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusNotFound, "League Not Found Or No Live Games At the Moment")
	} else if err != nil {
		log.Printf("getAllActive Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, games)
}

func (ld *LeaguesDelivery) get(c echo.Context) error {
	id := c.Param("id")
	league, err := (*ld.LeaguesHandler).Get(id)

	if err != nil {
		log.Printf("get league Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, league)
}
