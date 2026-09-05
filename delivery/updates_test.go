package delivery

import (
	"context"
	"dota_league/model"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type updatesServiceStub struct {
	offset, limit int
	filter        model.UpdateFilter
	err           error
}

func (s *updatesServiceStub) GetAll(ctx context.Context, offset, limit int, filters ...model.UpdateFilter) ([]model.Update, int64, error) {
	s.offset, s.limit = offset, limit
	if len(filters) > 0 {
		s.filter = filters[0]
	}
	if ctx.Err() != nil {
		return nil, 0, ctx.Err()
	}
	return nil, 0, s.err
}

func TestUpdatesEndpointPaginationAndFailures(t *testing.T) {
	e := echo.New()
	service := &updatesServiceStub{}
	NewUpdatesDelivery(e, service)
	for _, tc := range []struct {
		query         string
		offset, limit int
	}{
		{"", 0, 10}, {"?offset=20&limit=500", 20, 100}, {"?offset=-1&limit=0", 0, 10},
	} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/updates"+tc.query, nil))
		if rec.Code != 200 || service.offset != tc.offset || service.limit != tc.limit || !strings.Contains(rec.Body.String(), `"results":[]`) {
			t.Fatalf("query %s: %d %s", tc.query, rec.Code, rec.Body)
		}
	}
	for _, query := range []string{"?entity=unknown", "?days=-1", "?days=year"} {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/updates"+query, nil))
		if rec.Code != 400 {
			t.Fatalf("invalid filter %s: %d", query, rec.Code)
		}
	}
	recFilter := httptest.NewRecorder()
	e.ServeHTTP(recFilter, httptest.NewRequest(http.MethodGet, "/updates?search=Navi&entity=roster&days=7", nil))
	if recFilter.Code != 200 || service.filter != (model.UpdateFilter{Search: "Navi", Entity: "roster", Days: 7}) {
		t.Fatalf("filter not forwarded: %+v", service.filter)
	}
	service.err = errors.New("private database details")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/updates", nil))
	if rec.Code != http.StatusBadGateway || strings.Contains(rec.Body.String(), "private") {
		t.Fatalf("unsafe failure response: %d %s", rec.Code, rec.Body)
	}
}
