package delivery

import (
	e "dota_league/error"
	"log"
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
	teamsFromDB, totalCount, err := td.TeamsHandler.GetAll(meta.Offset, meta.Limit)
	if err != nil {
		log.Printf("getAll teams Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
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
	team, err := td.TeamsHandler.GetByID(id)
	if e.IsNotFound(err) {
		log.Printf("getOne team Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusNotFound, "Team Not Found")
	} else if err != nil {
		log.Printf("getOne team Delivery error: %+v,  message: %+v", err, e.ErrorMessage(err))
		return echo.NewHTTPError(http.StatusBadGateway, "Please try again later")
	}

	return c.JSON(http.StatusOK, team)
}
