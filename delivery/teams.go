package delivery

import (
	"dota_league/handler"

	"github.com/labstack/echo/v4"
)

// TeamsDelivery struct
type TeamsDelivery struct {
	TeamsHandler *handler.TeamsHandlerInterface
}

// NewTeamsDelivery adds routes to echo
func NewTeamsDelivery(e *echo.Echo, th *handler.TeamsHandlerInterface) {
	teamsDelivery := &TeamsDelivery{
		TeamsHandler: th,
	}

	e.GET("/teams", teamsDelivery.getAll)
	e.GET("/teams/:id", teamsDelivery.getOne)
}

func (td *TeamsDelivery) getAll(c echo.Context) error {

	return nil
}

func (td *TeamsDelivery) getOne(c echo.Context) error {

	return nil
}
