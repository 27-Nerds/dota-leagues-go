package delivery

import (
	"log"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
)

type response struct {
	Meta    *meta       `json:"meta"`
	Results interface{} `json:"results"`
}

type meta struct {
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

func newMeta(c *echo.Context) *meta {
	v, err := url.ParseQuery((*c).QueryString())
	if err != nil {
		log.Println("error")
		return nil
	}
	for id, thing := range v {
		log.Printf("id: %+v, t: %+v", id, thing)
	}

	log.Println((*c).QueryParam("filter[name]"))

	offset, err := strconv.Atoi((*c).QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}
	limit, err := strconv.Atoi((*c).QueryParam("limit"))
	if err != nil || limit < 0 {
		limit = 10
	}
	// dont allow to query more than 100 results in 1 query
	if limit > 100 {
		limit = 100
	}

	return &meta{
		Limit:  limit,
		Offset: offset,
	}
}
