package delivery

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestListFilterValidation(t *testing.T) {
	e := echo.New()
	for _, query := range []string{"sort=unknown", "order=DROP", "country=USA", "country=12", "active=maybe", "pro=maybe", "active_days=0", "active_days=3651"} {
		c := e.NewContext(httptest.NewRequest("GET", "/teams?"+query, nil), httptest.NewRecorder())
		if _, err := teamFilter(c); err == nil {
			t.Errorf("accepted %s", query)
		}
	}
	c := e.NewContext(httptest.NewRequest("GET", "/teams?country=ua&active=true&sort=name&order=asc", nil), httptest.NewRecorder())
	f, err := teamFilter(c)
	if err != nil || f.Country != "UA" || !f.Active || f.ActiveDays != 90 || f.Sort != "name" || f.Order != "asc" {
		t.Fatalf("unexpected filter: %+v %v", f, err)
	}
}

func TestProfessionalTeamFilter(t *testing.T) {
	e := echo.New()
	for _, value := range []string{"", "true", "false"} {
		c := e.NewContext(httptest.NewRequest("GET", "/teams?pro="+value, nil), httptest.NewRecorder())
		f, err := teamFilter(c)
		if err != nil {
			t.Fatal(err)
		}
		if value == "" {
			if f.Pro != nil {
				t.Fatal("omitted pro must include all teams")
			}
		} else if f.Pro == nil || *f.Pro != (value == "true") {
			t.Fatalf("incorrect pro filter for %q: %+v", value, f)
		}
	}
}
