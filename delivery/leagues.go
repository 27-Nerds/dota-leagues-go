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
}

// NewLeaguesDelivery adds routes to echo
func NewLeaguesDelivery(e *echo.Echo, lh *handler.LeaguesHandlerInterface) {
	leaguesDelivery := &LeaguesDelivery{
		LeaguesHandler: lh,
	}

	e.GET("/leagues", leaguesDelivery.getAllActive)
}

func (ld *LeaguesDelivery) getAllActive(c echo.Context) error {
	leagues, err := (*ld.LeaguesHandler).GetAllActive()
	if err != nil {
		log.Printf("getAllActive Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, leagues)
}
