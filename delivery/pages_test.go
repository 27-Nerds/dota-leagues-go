package delivery

import (
	"context"
	appError "dota_league/error"
	"dota_league/model"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type pageLeagues struct{ err error }

func (f pageLeagues) GetAllActive(ctx context.Context, offset, limit int) ([]model.LeagueDetails, int64, error) {
	rows := []model.LeagueDetails{{ID: 1, Name: "Test League"}}
	if offset > 1000 {
		rows = nil
	}
	return rows, 1001, f.err
}
func (f pageLeagues) GetByID(ctx context.Context, id string) (*model.LeagueDetails, error) {
	if id == "404" {
		return nil, &appError.Error{Code: appError.ENOTFOUND}
	}
	return &model.LeagueDetails{ID: 1, Name: "Test <League>", Description: `<script>alert("x")</script>`}, f.err
}
func (f pageLeagues) GetSeries(ctx context.Context, id string, offset, limit int) ([]model.LeagueSeries, int64, error) {
	rows := []model.LeagueSeries{{TeamID1: 2, TeamName1: "Test Team", MatchIDs: []string{"123"}}}
	return rows, 101, f.err
}

type pageTeams struct{}

func (pageTeams) GetAll(ctx context.Context, offset, limit int) ([]model.Team, int64, error) {
	rows := []model.Team{{ID: 2, Name: "Test Team"}}
	return rows, 1, nil
}
func (pageTeams) GetByID(ctx context.Context, id string) (*model.Team, error) {
	return &model.Team{ID: 2, Name: "Test Team", Wins: 5}, nil
}

type pageMatches struct{}

func (pageMatches) Get(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
	return &model.MatchMinimal{MatchID: matchID, Tourney: model.MatchTourney{RadiantTeamName: "Team A", DireTeamName: "Team B"}}, nil
}

type pageStandings struct{}

func (pageStandings) Get(ctx context.Context) (*model.DPCStandings, error) {
	return &model.DPCStandings{}, nil
}

func pageServer(t *testing.T, serviceErr error) *echo.Echo {
	t.Helper()
	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte(`<!doctype html><html><head><title>Dota 2 leagues</title><!--seo-head--></head><body><div id="app"><!--seo-body--></div><script type="module" src="/build/bundle.js"></script></body></html>`), 0600); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	e.Static("/", t.TempDir())
	e.GET("/leagues", func(c echo.Context) error { return c.JSON(200, map[string]string{"api": "unchanged"}) })
	if err := NewPagesDelivery(e, pageLeagues{serviceErr}, pageTeams{}, pageMatches{}, pageStandings{}, path, "https://example.com"); err != nil {
		t.Fatal(err)
	}
	return e
}
func getPage(e *echo.Echo, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}
func TestCrawlablePages(t *testing.T) {
	e := pageServer(t, nil)
	for _, tc := range []struct{ path, text string }{
		{"/", `href="/league/1"`},
		{"/?offset=100", `https://example.com/?offset=100`},
		{"/league/1", `href="/match/1/123"`},
		{"/team/2", "5 wins"},
		{"/match/1/123", "Team A"},
		{"/dpc", "No standings available"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			res := getPage(e, tc.path)
			if res.Code != 200 {
				t.Fatalf("status %d: %s", res.Code, res.Body)
			}
			html := res.Body.String()
			for _, want := range []string{tc.text, `rel="canonical"`, `name="description"`, `id="app"`, `src="/build/bundle.js"`} {
				if !strings.Contains(html, want) {
					t.Errorf("missing %s in %s", want, html)
				}
			}
			if strings.Count(html, "<title>") != 1 || strings.Contains(html, "<!--seo-") {
				t.Error("unreplaced HTML shell")
			}
		})
	}
	escaped := getPage(e, "/league/1").Body.String()
	if strings.Contains(escaped, "<script>alert") || !strings.Contains(escaped, "&lt;script&gt;") {
		t.Fatal("unescaped service content")
	}
	if !strings.Contains(getPage(e, "/leagues").Body.String(), "unchanged") {
		t.Fatal("API route shadowed")
	}
}
func TestPageFailures(t *testing.T) {
	e := pageServer(t, nil)
	for _, path := range []string{"/league/404", "/league/abc", "/team/0", "/match/1/no", "/unknown", "/?offset=-1", "/?offset=1"} {
		if res := getPage(e, path); res.Code != 404 {
			t.Errorf("%s: got %d", path, res.Code)
		}
	}
	if res := getPage(pageServer(t, errors.New("database down")), "/league/1"); res.Code != 503 {
		t.Errorf("outage returned %d", res.Code)
	}
}
func TestSitemaps(t *testing.T) {
	e := pageServer(t, nil)
	for _, tc := range []struct{ path, root, want string }{
		{"/sitemap.xml", "sitemapindex", "https://example.com/sitemaps/leagues/2"},
		{"/sitemaps/pages/1", "urlset", "https://example.com/dpc"},
		{"/sitemaps/leagues/1", "urlset", "https://example.com/league/1"},
		{"/sitemaps/teams/1", "urlset", "https://example.com/team/2"},
	} {
		res := getPage(e, tc.path)
		var doc sitemapDocument
		if res.Code != 200 {
			t.Fatalf("%s: %d", tc.path, res.Code)
		}
		if err := xml.Unmarshal(res.Body.Bytes(), &doc); err != nil {
			t.Fatal(err)
		}
		if doc.XMLName.Local != tc.root || !strings.Contains(res.Body.String(), tc.want) {
			t.Fatalf("invalid sitemap: %s", res.Body)
		}
	}
	if res := getPage(e, "/robots.txt"); !strings.Contains(res.Body.String(), "Sitemap: https://example.com/sitemap.xml") {
		t.Fatal(res.Body)
	}
	for _, path := range []string{"/sitemaps/pages/2", "/sitemaps/leagues/0", "/sitemaps/unknown/1"} {
		if res := getPage(e, path); res.Code != 404 {
			t.Errorf("%s: %d", path, res.Code)
		}
	}
}
