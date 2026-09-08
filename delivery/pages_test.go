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

func (f pageLeagues) GetAll(ctx context.Context, offset, limit int, filter model.LeagueFilter) ([]model.LeagueDetails, int64, error) {
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

func (pageTeams) GetAll(ctx context.Context, offset, limit int, filter model.TeamFilter) ([]model.Team, int64, error) {
	rows := []model.Team{{ID: 2, Name: "Test Team"}}
	if offset > 100 {
		rows = nil
	}
	return rows, 101, nil
}
func (pageTeams) GetByID(ctx context.Context, id string) (*model.Team, error) {
	return &model.Team{ID: 2, Name: "Test Team", Wins: 5}, nil
}

type pagePlayers struct{}

func (pagePlayers) GetAll(ctx context.Context, offset, limit int, filter model.PlayerFilter) ([]model.Player, int64, error) {
	if offset > 100 {
		return nil, 101, nil
	}
	return []model.Player{{ID: 9, Name: "Pro <One>"}}, 101, nil
}
func (pagePlayers) GetByID(ctx context.Context, id string) (*model.Player, error) {
	if id == "404" {
		return nil, &appError.Error{Code: appError.ENOTFOUND}
	}
	return &model.Player{ID: 9, Name: "Pro <One>", RealName: "Real Name", TeamID: 2, TeamName: "Test Team", SteamName: "steam one", SteamLocation: "Kyiv", SteamStatus: "ok", SteamPrivacy: "private", SteamUpdatedAt: 1788609600000, SteamPublicAt: 1788609600000, RosterTeam: &model.PlayerRosterTeam{ID: 3, Name: "Roster Team"}, TeamHistory: []model.PlayerTeamEntry{{TeamID: 4, TeamName: "Old Team", StartTimestamp: 1788609600}}, Results: []model.PlayerResult{{LeagueID: 1, Placement: 3, LeagueName: "Test League"}}}, nil
}
func (pagePlayers) GetSitemapPlayers(ctx context.Context, offset, limit int) ([]int, int64, error) {
	if offset > 0 {
		return nil, 1, nil
	}
	return []int{9}, 1, nil
}

type pageMatches struct{}

func (pageMatches) Get(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
	return &model.MatchMinimal{MatchID: matchID, Tourney: model.MatchTourney{RadiantTeamName: "Team A", DireTeamName: "Team B"}}, nil
}

type pageStandings struct{}

func (pageStandings) Get(ctx context.Context) (*model.DPCStandings, error) {
	return &model.DPCStandings{}, nil
}

func pageServer(t *testing.T, serviceErr error, analytics ...string) *echo.Echo {
	analyticsID := ""
	if len(analytics) > 0 {
		analyticsID = analytics[0]
	}
	t.Helper()
	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte(`<!doctype html><html><head><title>Dota 2 leagues</title><!--seo-head--></head><body><div id="app"><!--seo-body--></div><script type="module" src="/build/bundle.js"></script></body></html>`), 0600); err != nil {
		t.Fatal(err)
	}
	e := echo.New()
	e.Static("/", t.TempDir())
	e.GET("/leagues", func(c echo.Context) error { return c.JSON(200, map[string]string{"api": "unchanged"}) })
	if err := NewPagesDelivery(e, pageLeagues{serviceErr}, pageTeams{}, pagePlayers{}, pageMatches{}, pageStandings{}, pageUpdates{serviceErr}, pageMatchIndex{}, path, "https://example.com", analyticsID); err != nil {
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
		{"/player/9", "Pro &lt;One&gt;"},
		{"/player", `href="/player/9"`},
		{"/player", `href="/player?offset=100"`},
		{"/player/9", "Steam profile is private; Steam still shows the name and avatar, other details were last seen public on 5 Sep 2026"},
		{"/player/9", `href="/team/2"`},
		{"/player/9", "Test League: place 3"},
		{"/player/9", `href="/team/3"`},
		{"/player/9", `href="/team/4">Old Team since 5 Sep 2026<`},
		{"/player/9", `"@type":"Person"`},
		{"/team", `href="/team/2"`},
		{"/team?offset=100", `https://example.com/team?offset=100`},
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
	for _, path := range []string{"/league/404", "/league/abc", "/team/0", "/player/404", "/match/1/no", "/unknown", "/?offset=-1", "/?offset=1"} {
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
		{"/sitemaps/pages/1", "urlset", "https://example.com/team"},
		{"/sitemaps/leagues/1", "urlset", "https://example.com/league/1"},
		{"/sitemaps/teams/1", "urlset", "https://example.com/team/2"},
		{"/sitemaps/players/1", "urlset", "https://example.com/player/9"},
		{"/sitemaps/pages/1", "urlset", "https://example.com/player"},
		{"/sitemap.xml", "sitemapindex", "https://example.com/sitemaps/players/1"},
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

func TestTeamDirectoryPagination(t *testing.T) {
	e := pageServer(t, nil)
	if body := getPage(e, "/team?country=UA&active=true&sort=name").Body.String(); !strings.Contains(body, `href="/team?active=true&amp;country=UA&amp;offset=100&amp;sort=name"`) {
		t.Fatalf("pagination lost filters: %s", body)
	}
	if body := getPage(e, "/team").Body.String(); !strings.Contains(body, `href="/team?offset=100"`) {
		t.Fatalf("missing next page: %s", body)
	}
	if body := getPage(e, "/team?offset=100").Body.String(); !strings.Contains(body, `href="/team"`) {
		t.Fatalf("missing previous page: %s", body)
	}
	if res := getPage(e, "/team?offset=200"); res.Code != http.StatusNotFound {
		t.Fatalf("empty directory page returned %d", res.Code)
	}
	NewTeamsDelivery(e, pageTeams{})
	if res := getPage(e, "/teams"); res.Code != http.StatusOK || !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("teams API shadowed: %d %s", res.Code, res.Body)
	}
}

func documentRequest(e *echo.Echo, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHTMLDocumentErrors(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		serviceErr error
		status     int
		title      string
	}{
		{"unknown page", "/unknown", nil, 404, "This page is off the map"},
		{"missing league", "/league/404", nil, 404, "This page is off the map"},
		{"invalid team", "/team/0", nil, 404, "This page is off the map"},
		{"outage", "/league/1", errors.New("secret database error"), 503, "We couldn’t load this page"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := pageServer(t, tc.serviceErr)
			res := documentRequest(e, http.MethodGet, tc.path)
			if res.Code != tc.status || !strings.Contains(res.Header().Get("Content-Type"), "text/html") {
				t.Fatalf("got %d %s", res.Code, res.Body)
			}
			for _, want := range []string{tc.title, `name="robots" content="noindex"`, `href="/team"`, `href="/build/index.css"`} {
				if !strings.Contains(res.Body.String(), want) {
					t.Errorf("missing %q: %s", want, res.Body)
				}
			}
			if strings.Contains(res.Body.String(), "secret database") || strings.Contains(res.Body.String(), "bundle.js") {
				t.Fatal("error page leaks internals or replaces fallback with app")
			}
			if tc.status == 503 && !strings.Contains(res.Body.String(), "Try again") {
				t.Fatal("outage page has no retry")
			}
			headStatus := tc.status
			if tc.path == "/unknown" {
				headStatus = http.StatusMethodNotAllowed
			}
			if res := documentRequest(e, http.MethodHead, tc.path); res.Code != headStatus || res.Body.Len() != 0 {
				t.Fatalf("HEAD: %d %s", res.Code, res.Body)
			}
		})
	}
}

func TestHTMLHandlerPreservesAPIErrors(t *testing.T) {
	e := pageServer(t, nil)
	e.GET("/teams", func(c echo.Context) error { return echo.NewHTTPError(http.StatusServiceUnavailable, "API unavailable") })
	e.GET("/leagues/:id", func(c echo.Context) error { return echo.NewHTTPError(http.StatusNotFound, "League missing") })
	for _, path := range []string{"/teams", "/leagues/404", "/leagues/missing/nested", "/teams/missing", "/dpc/standings", "/build/missing.js", "/assets/missing", "/missing.png", "/sitemaps/unknown/1"} {
		res := documentRequest(e, http.MethodGet, path)
		if res.Code != 404 && res.Code != 503 {
			t.Errorf("%s: status %d", path, res.Code)
		}
		if !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
			t.Errorf("%s error became HTML: %s", path, res.Body)
		}
	}
	if res := getPage(e, "/unknown"); !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
		t.Fatal("non-document error became HTML")
	}
}

func TestSchedulePageLinks(t *testing.T) {
	e := pageServer(t, nil)
	for _, path := range []string{"/league/1", "/league/1?offset=20", "/league/1?offset=100"} {
		res := getPage(e, path)
		if res.Code != http.StatusOK {
			t.Fatalf("%s: got %d", path, res.Code)
		}
	}
	first := getPage(e, "/league/1").Body.String()
	if !strings.Contains(first, `href="/league/1?offset=20"`) {
		t.Fatal("schedule next link must advance by 20")
	}
	second := getPage(e, "/league/1?offset=20").Body.String()
	for _, want := range []string{`href="/league/1?offset=0"`, `href="/league/1?offset=40"`, `href="https://example.com/league/1?offset=20"`} {
		if !strings.Contains(second, want) {
			t.Errorf("missing %s", want)
		}
	}
	for _, path := range []string{"/league/1?offset=19", "/?offset=20", "/team?offset=20"} {
		if res := getPage(e, path); res.Code != http.StatusNotFound {
			t.Errorf("%s: got %d", path, res.Code)
		}
	}
}

type pageUpdates struct{ err error }

func (p pageUpdates) GetAll(_ context.Context, offset, limit int, _ ...model.UpdateFilter) ([]model.Update, int64, error) {
	if offset >= 40 {
		return nil, 21, p.err
	}
	return []model.Update{{ID: "event", Entity: "team", EntityID: 7, Name: "Team <A>", Action: "updated", CreatedAt: 1788609600000, Changes: []model.UpdateChange{{Field: "wins", Before: 1, After: 2}}}}, 21, p.err
}
func TestUpdatesPages(t *testing.T) {
	e := pageServer(t, nil)
	NewUpdatesDelivery(e, pageUpdates{})
	for _, tc := range []struct{ path, want string }{
		{"/activity", `href="/activity?offset=20"`},
		{"/activity?offset=20", `href="https://example.com/activity?offset=20"`},
		{"/sitemaps/pages/1", "https://example.com/activity"},
	} {
		res := getPage(e, tc.path)
		if res.Code != 200 || !strings.Contains(res.Body.String(), tc.want) {
			t.Fatalf("%s: %d %s", tc.path, res.Code, res.Body)
		}
	}
	body := getPage(e, "/activity").Body.String()
	for _, want := range []string{`href="/team/7"`, "Team &lt;A&gt;", "wins: 1 → 2"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %s", want)
		}
	}
	for _, path := range []string{"/activity?offset=1", "/activity?offset=40"} {
		if res := getPage(e, path); res.Code != 404 {
			t.Errorf("%s: %d", path, res.Code)
		}
	}
	if res := documentRequest(pageServer(t, errors.New("offline")), http.MethodGet, "/activity"); res.Code != 503 || !strings.Contains(res.Header().Get("Content-Type"), "text/html") {
		t.Fatal("activity outage missing HTML error")
	}
	if res := documentRequest(e, http.MethodGet, "/updates?offset=0&limit=20"); res.Code != 200 || !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
		t.Fatal("updates API shadowed")
	}
}

type pageMatchIndex struct{}

func (pageMatchIndex) GetSitemapMatches(ctx context.Context, offset, limit int) ([]model.MatchReference, int64, error) {
	if offset > 0 {
		return nil, 1, nil
	}
	return []model.MatchReference{{LeagueID: 1, MatchID: "123"}}, 1, nil
}

func TestIndexingPolicyAndMatchSitemaps(t *testing.T) {
	e := pageServer(t, nil)
	for _, path := range []string{"/?search=Test", "/?status=all", "/?tier=2&offset=100", "/team?country=US", "/team?sort=wins"} {
		r := getPage(e, path)
		if r.Code != 200 || !strings.Contains(r.Body.String(), `name="robots" content="noindex,follow"`) {
			t.Fatalf("filtered page %s: %d %s", path, r.Code, r.Body)
		}
	}
	for _, path := range []string{"/", "/?offset=100", "/team", "/team?offset=100", "/league/1?data=1"} {
		r := getPage(e, path)
		if r.Code != 200 || strings.Contains(r.Body.String(), "noindex") {
			t.Fatalf("indexable page %s: %d %s", path, r.Code, r.Body)
		}
	}
	for _, tc := range []struct{ path, want string }{{"/sitemap.xml", "https://example.com/sitemaps/matches/1"}, {"/sitemaps/matches/1", "https://example.com/match/1/123"}} {
		r := getPage(e, tc.path)
		if r.Code != 200 || !strings.Contains(r.Body.String(), tc.want) {
			t.Fatalf("sitemap %s: %d %s", tc.path, r.Code, r.Body)
		}
	}
	if r := getPage(e, "/sitemaps/matches/2"); r.Code != 404 {
		t.Fatalf("empty match sitemap: %d", r.Code)
	}
}

func TestMatchPageUsesOutcomeNotKillsAndNamesHeroes(t *testing.T) {
	match := &model.MatchMinimal{MatchID: "123", MatchOutcome: 3, RadiantScore: 40, DireScore: 12, StartTime: 1698513226, Tourney: model.MatchTourney{RadiantTeamID: 1, RadiantTeamName: "A", DireTeamID: 2, DireTeamName: "B"}, Players: []model.MatchMinimalPlayer{{AccountID: 42, HeroID: 74}, {HeroID: 999}}}
	var page pageContent
	populateMatchPage(&page, match, "1")
	if !strings.Contains(page.Description, "B won.") || !strings.Contains(page.Description, "Kills: 40 to 12") {
		t.Fatal(page.Description)
	}
	body := strings.Join(page.Paragraphs, "\n")
	for _, want := range []string{"28 Oct 2023", "Player #42 (Invoker, Radiant)", "Anonymous player (Unknown hero"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %s: %s", want, body)
		}
	}
	match.MatchOutcome = 0
	populateMatchPage(&page, match, "1")
	if !strings.Contains(page.Description, "Result unavailable") {
		t.Fatal(page.Description)
	}
}

func TestOptionalGoogleAnalytics(t *testing.T) {
	for _, id := range []string{"", "   "} {
		html := getPage(pageServer(t, nil, id), "/").Body.String()
		if strings.Contains(html, "googletagmanager") || strings.Contains(html, "gtag(") || strings.Contains(html, "dataLayer") {
			t.Fatal("analytics loaded while disabled")
		}
	}
	e := pageServer(t, nil, " G-ABC1234567 ")
	for _, path := range []string{"/", "/league/1", "/team/2", "/match/1/123", "/activity"} {
		html := getPage(e, path).Body.String()
		if strings.Count(html, `src="https://www.googletagmanager.com/gtag/js?id=G-ABC1234567"`) != 1 || strings.Count(html, `gtag('config', "G-ABC1234567");`) != 1 {
			t.Fatalf("missing/duplicate analytics: %s", html)
		}
	}
	for _, id := range []string{"UA-123-1", "GTM-ABC", "G-", "G-ABC\"</script>", "G-ABC xyz"} {
		err := NewPagesDelivery(echo.New(), nil, nil, nil, nil, nil, nil, nil, "unused", "https://example.com", id)
		if err == nil || !strings.Contains(err.Error(), "GA_MEASUREMENT_ID") {
			t.Fatalf("accepted bad ID %q: %v", id, err)
		}
	}
}

func TestPlayerTeamSourcesAvoidActiveTeamClaims(t *testing.T) {
	body := getPage(pageServer(t, nil), "/player/9").Body.String()
	for _, want := range []string{"DPC-listed team: Test Team", "Team listings differ.", "Listed on team roster: Roster Team"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "playing for") {
		t.Error("stored listing described as active membership")
	}
}
