package delivery

import (
	"dota_league/model"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func listSort(c echo.Context, allowed ...string) (string, string, error) {
	sort, order := c.QueryParam("sort"), c.QueryParam("order")
	if sort != "" && sort != "recommended" && !slices.Contains(allowed, sort) {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "Invalid sort")
	}
	if order != "" && order != "asc" && order != "desc" {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "Invalid sort order")
	}
	return sort, order, nil
}

func teamFilter(c echo.Context) (model.TeamFilter, error) {
	f := model.TeamFilter{Search: c.QueryParam("search"), Country: strings.ToUpper(strings.TrimSpace(c.QueryParam("country"))), ActiveDays: 90}
	var err error
	f.Sort, f.Order, err = listSort(c, "name", "wins", "activity")
	if err != nil {
		return f, err
	}
	if f.Country != "" && (len(f.Country) != 2 || f.Country[0] < 'A' || f.Country[0] > 'Z' || f.Country[1] < 'A' || f.Country[1] > 'Z') {
		return f, echo.NewHTTPError(http.StatusBadRequest, "Country must be a two-letter code")
	}
	if value := c.QueryParam("pro"); value != "" {
		pro, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return f, echo.NewHTTPError(http.StatusBadRequest, "Invalid professional filter")
		}
		f.Pro = &pro
	}
	if value := c.QueryParam("active"); value != "" {
		f.Active, err = strconv.ParseBool(value)
		if err != nil {
			return f, echo.NewHTTPError(http.StatusBadRequest, "Invalid active filter")
		}
	}
	if value := c.QueryParam("active_days"); value != "" {
		f.ActiveDays, err = strconv.Atoi(value)
		if err != nil || f.ActiveDays < 1 || f.ActiveDays > 3650 {
			return f, echo.NewHTTPError(http.StatusBadRequest, "Active days must be between 1 and 3650")
		}
	}
	return f, nil
}

func leagueFilter(c echo.Context) (model.LeagueFilter, error) {
	f := model.LeagueFilter{Status: c.QueryParam("status"), Search: c.QueryParam("search"), LiveOnly: c.QueryParam("live") == "true"}
	var err error
	f.Sort, f.Order, err = listSort(c, "name", "start_date", "end_date", "prize_pool", "tier")
	if err != nil {
		return f, err
	}
	if f.Status != "" && f.Status != "active" && f.Status != "completed" && f.Status != "all" {
		return f, echo.NewHTTPError(http.StatusBadRequest, "Invalid tournament status")
	}
	for key, target := range map[string]**int{"tier": &f.Tier, "region": &f.Region} {
		if value := c.QueryParam(key); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 {
				return f, echo.NewHTTPError(http.StatusBadRequest, "Invalid "+key)
			}
			*target = &n
		}
	}
	return f, nil
}
