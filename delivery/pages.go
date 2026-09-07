package delivery

import (
	"bytes"
	appError "dota_league/error"
	"dota_league/model"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type pageLink struct{ URL, Label string }
type pageContent struct {
	Title, Description, Canonical string
	NoIndex                       bool
	Entity                        schemaNode
	StructuredData                template.JS
	AnalyticsID                   string
	Links                         []pageLink
	Paragraphs                    []string
}

// NewPagesDelivery serves crawlable HTML using the same services as the JSON API.
// The Svelte app progressively replaces this initial content in the browser.
func NewPagesDelivery(e *echo.Echo, leagues LeaguesService, teams TeamsService, players PlayersService, matches MatchService, standings StandingsService, updates UpdatesService, matchIndex MatchSitemapService, indexPath, siteURL, analyticsID string) error {
	analyticsID = strings.TrimSpace(analyticsID)
	if analyticsID != "" && !regexp.MustCompile(`^G-[A-Z0-9]+$`).MatchString(analyticsID) {
		return fmt.Errorf("GA_MEASUREMENT_ID must be a GA4 measurement ID beginning with G-")
	}
	origin, err := url.Parse(siteURL)
	if err != nil || origin.Host == "" || (origin.Scheme != "https" && origin.Scheme != "http") || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") {
		return fmt.Errorf("SITE_URL must be an absolute http(s) origin")
	}
	p := &pagesDelivery{leagues: leagues, teams: teams, players: players, matches: matches, standings: standings, updates: updates, matchIndex: matchIndex, indexPath: indexPath, siteURL: strings.TrimRight(siteURL, "/"), analyticsID: analyticsID}
	for _, route := range []string{"/", "/team", "/dpc", "/activity", "/league/:id", "/team/:id", "/player/:id", "/match/:leagueID/:matchID"} {
		e.GET(route, p.page)
		e.HEAD(route, p.page)
	}
	previousErrorHandler := e.HTTPErrorHandler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if !c.Response().Committed && isHTMLPageRequest(c) {
			var httpError *echo.HTTPError
			status := http.StatusInternalServerError
			if errors.As(err, &httpError) {
				status = httpError.Code
			}
			if status == http.StatusNotFound || status >= 500 {
				if renderErr := p.errorPage(c, status); renderErr == nil {
					return
				}
			}
		}
		previousErrorHandler(err, c)
	}

	e.GET("/robots.txt", func(c echo.Context) error {
		return c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: "+p.siteURL+"/sitemap.xml\n")
	})
	e.GET("/sitemap.xml", p.sitemap)
	e.GET("/sitemaps/:kind/:page", p.sitemap)
	return nil
}

type pagesDelivery struct {
	leagues     LeaguesService
	teams       TeamsService
	players     PlayersService
	matches     MatchService
	standings   StandingsService
	updates     UpdatesService
	matchIndex  MatchSitemapService
	indexPath   string
	siteURL     string
	analyticsID string
}

var pageHead = template.Must(template.New("head").Parse(`<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<link rel="canonical" href="{{.Canonical}}">
{{if .NoIndex}}<meta name="robots" content="noindex,follow">{{end}}
<meta property="og:type" content="website">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:url" content="{{.Canonical}}">
<script type="application/ld+json">{{.StructuredData}}</script>
{{if .AnalyticsID}}<script async src="https://www.googletagmanager.com/gtag/js?id={{.AnalyticsID}}"></script>
<script>
window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', {{.AnalyticsID}});
</script>{{end}}`))

var pageBody = template.Must(template.New("body").Parse(`<header><a href="/">Dota 2 Leagues</a> · <a href="/team">Teams</a> · <a href="/dpc">DPC Standings</a> · <a href="/activity">Updates</a></header>
<main><h1>{{.Title}}</h1><p>{{.Description}}</p>
{{range .Paragraphs}}<p>{{.}}</p>{{end}}
<ul>{{range .Links}}<li><a href="{{.URL}}">{{.Label}}</a></li>{{end}}</ul></main>`))

func positiveID(s string) bool {
	n, err := strconv.ParseUint(s, 10, 64)
	return err == nil && n > 0 && strconv.FormatUint(n, 10) == s
}

func pageOffset(c echo.Context) (int, error) {
	s := c.QueryParam("offset")
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	pageSize := 100
	if c.Path() == "/league/:id" || c.Path() == "/activity" {
		pageSize = 20
	}
	if err != nil || n < 0 || n%pageSize != 0 || n > 10000000 {
		return 0, echo.NewHTTPError(http.StatusNotFound)
	}
	return n, nil
}

func (p *pagesDelivery) page(c echo.Context) error {
	for _, id := range c.ParamValues() {
		if !positiveID(id) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
	}
	offset, err := pageOffset(c)
	if err != nil {
		return err
	}
	data := pageContent{Title: "Dota 2 Leagues", Description: "Explore Dota 2 tournaments, live games, match results and teams.", Canonical: p.siteURL + c.Request().URL.Path}
	if offset > 0 {
		data.Canonical += "?offset=" + strconv.Itoa(offset)
	}
	if c.Path() == "/" || c.Path() == "/team" {
		for key, values := range c.QueryParams() {
			if key == "offset" {
				continue
			}
			for _, value := range values {
				if value != "" && strings.Contains("|search|status|tier|region|live|country|pro|active|active_days|sort|order|", "|"+key+"|") {
					data.NoIndex = true
				}
			}
		}
	}
	switch c.Path() {
	case "/activity":
		data.Title = "Updates | Dota 2 Leagues"
		data.Description = "Recorded tournament changes, team records, roster moves, and newly added players. Newest first."
		var rows []model.Update
		var total int64
		rows, total, err = p.updates.GetAll(c.Request().Context(), offset, 20)
		if err == nil {
			if offset > 0 && len(rows) == 0 {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			if len(rows) == 0 {
				data.Paragraphs = append(data.Paragraphs, "No updates recorded yet. Earlier changes are not included.")
			}
			for _, row := range rows {
				action := "updated"
				if row.Action == "created" {
					action = "added"
				}
				data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("%s UTC: %s %s — %s", time.UnixMilli(row.CreatedAt).UTC().Format("2 Jan 2006 15:04"), row.Entity, action, row.Name))
				if row.Entity == "player" && row.Team != nil && row.Team.ID > 0 {
					label := "Team"
					name := row.Team.Name
					if name == "" {
						name = fmt.Sprintf("Team #%d", row.Team.ID)
					}
					data.Paragraphs = append(data.Paragraphs, label+": "+name)
					data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", row.Team.ID), name})
				}
				for _, change := range row.Changes {
					data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("%s: %v → %v", strings.ReplaceAll(change.Field, "_", " "), change.Before, change.After))
				}
				if row.EntityID > 0 {
					switch row.Entity {
					case "tournament":
						data.Links = append(data.Links, pageLink{fmt.Sprintf("/league/%d", row.EntityID), row.Name})
					case "team", "roster":
						data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", row.EntityID), row.Name})
					case "player":
						data.Links = append(data.Links, pageLink{fmt.Sprintf("/player/%d", row.EntityID), row.Name})
					}
				}
			}
			addPagination(&data, "/activity", offset, total, 20)
		}
	case "/":
		filter, filterErr := leagueFilter(c)
		if filterErr != nil {
			return filterErr
		}
		data.Canonical = p.siteURL + listPageURL(c, offset)
		var rows []model.LeagueDetails
		var total int64
		rows, total, err = p.leagues.GetAll(c.Request().Context(), offset, 100, filter)
		if err == nil {
			for _, l := range rows {
				data.Links = append(data.Links, pageLink{fmt.Sprintf("/league/%d", l.ID), l.Name})
			}
			if offset > 0 && len(rows) == 0 {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			addListPagination(&data, c, offset, total)
		}
	case "/team":
		filter, filterErr := teamFilter(c)
		if filterErr != nil {
			return filterErr
		}
		data.Canonical = p.siteURL + listPageURL(c, offset)
		data.Title = "Dota 2 Teams"
		data.Description = "Browse Dota 2 teams, player rosters and tournament results."
		var rows []model.Team
		var total int64
		rows, total, err = p.teams.GetAll(c.Request().Context(), offset, 100, filter)
		if err == nil {
			for _, team := range rows {
				name := team.Name
				if name == "" {
					name = fmt.Sprintf("Team #%d", team.ID)
				}
				data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", team.ID), name})
			}
			if len(rows) == 0 {
				if offset > 0 {
					return echo.NewHTTPError(http.StatusNotFound)
				}
				data.Paragraphs = append(data.Paragraphs, "No teams available yet.")
			}
			addListPagination(&data, c, offset, total)
		}
	case "/league/:id":
		var league *model.LeagueDetails
		league, err = p.leagues.GetByID(c.Request().Context(), c.Param("id"))
		if err == nil && league == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err == nil {
			data.Entity = leagueSchema(league, data.Canonical)
			data.Title = league.Name + " | Dota 2 Leagues"
			data.Description = league.Description
			if data.Description == "" {
				data.Description = league.Name + " schedule, live games and match results."
			}
			data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("Prize pool: $%d", league.TotalPrizePool))
			var series []model.LeagueSeries
			var total int64
			series, total, err = p.leagues.GetSeries(c.Request().Context(), c.Param("id"), offset, 20)
			if appError.IsNotFound(err) && offset == 0 {
				err = nil
			}
			if err == nil {
				for _, s := range series {
					for _, team := range []struct {
						id   int
						name string
					}{{s.TeamID1, s.TeamName1}, {s.TeamID2, s.TeamName2}} {
						if team.id > 0 {
							name := team.name
							if name == "" {
								name = fmt.Sprintf("Team #%d", team.id)
							}
							data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", team.id), name})
						}
					}
					for _, match := range s.MatchIDs {
						if positiveID(match) {
							data.Links = append(data.Links, pageLink{fmt.Sprintf("/match/%d/%s", league.ID, match), "Match #" + match})
						}
					}
				}
				if offset > 0 && len(series) == 0 {
					return echo.NewHTTPError(http.StatusNotFound)
				}
				addPagination(&data, c.Request().URL.Path, offset, total, 20)
			}
		}
	case "/team/:id":
		var team *model.Team
		team, err = p.teams.GetByID(c.Request().Context(), c.Param("id"))
		if err == nil && team == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err == nil {
			if strings.TrimSpace(team.Name) != "" {
				data.Entity = sportsEntity("SportsTeam", team.Name, data.Canonical, strconv.Itoa(team.ID))
			}
			data.Title = team.Name + " | Dota 2 Teams"
			data.Description = fmt.Sprintf("%s roster and tournament results. %d wins and %d losses.", team.Name, team.Wins, team.Losses)
			for _, member := range team.Members {
				name := member.ProName
				if name == "" {
					name = member.PlayerName
				}
				if member.AccountID > 0 {
					if name == "" {
						name = fmt.Sprintf("Player #%d", member.AccountID)
					}
					data.Links = append(data.Links, pageLink{fmt.Sprintf("/player/%d", member.AccountID), name})
				} else if name != "" {
					data.Paragraphs = append(data.Paragraphs, name)
				}
			}
			for _, result := range team.DpcResults {
				data.Links = append(data.Links, pageLink{fmt.Sprintf("/league/%d", result.LeagueID), fmt.Sprintf("League #%d: place %d", result.LeagueID, result.Standing)})
			}
		}
	case "/player/:id":
		var player *model.Player
		player, err = p.players.GetByID(c.Request().Context(), c.Param("id"))
		if err == nil && (player == nil || player.ID <= 0) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err == nil {
			populatePlayerPage(&data, player)
		}
	case "/match/:leagueID/:matchID":
		leagueID, parseErr := strconv.Atoi(c.Param("leagueID"))
		if parseErr != nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		var match *model.MatchMinimal
		match, err = p.matches.Get(c.Request().Context(), leagueID, c.Param("matchID"))
		if err == nil && (match == nil || match.MatchID == "") {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err == nil {
			populateMatchPage(&data, match, c.Param("leagueID"))
		}
	case "/dpc":
		data.Title = "DPC Standings | Dota 2 Leagues"
		data.Description = "Official Dota Pro Circuit standings and league results."
		var standings *model.DPCStandings
		standings, err = p.standings.Get(c.Request().Context())
		if err == nil && standings != nil {
			for _, rows := range [][]any{standings.Results, standings.Standings, standings.MajorWildcardStandings, standings.MajorGroupStandings, standings.MajorPlayoffStandings} {
				for _, row := range rows {
					data.Paragraphs = append(data.Paragraphs, fmt.Sprint(row))
				}
			}
			if len(data.Paragraphs) == 0 {
				data.Paragraphs = append(data.Paragraphs, "No standings available for the current season.")
			}
		}
	}
	if err != nil {
		if appError.IsNotFound(err) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Please try again later")
	}
	if c.Path() != "/" && c.Path() != "/team" && c.Path() != "/league/:id" && c.Path() != "/activity" {
		data.Canonical = p.siteURL + c.Request().URL.Path
	}
	shell, err := os.ReadFile(p.indexPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Frontend build unavailable")
	}
	if !bytes.Contains(shell, []byte("<!--seo-head-->")) || !bytes.Contains(shell, []byte("<!--seo-body-->")) {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Frontend build needs updating")
	}
	data.AnalyticsID = p.analyticsID
	data.StructuredData, err = structuredData(data, p.siteURL, c.Path(), c.Param("leagueID"))
	if err != nil {
		return err
	}
	var head, body bytes.Buffer
	if err = pageHead.Execute(&head, data); err != nil {
		return err
	}
	if err = pageBody.Execute(&body, data); err != nil {
		return err
	}
	html := strings.Replace(string(shell), "<title>Dota 2 leagues</title>", "", 1)
	html = strings.Replace(html, "<!--seo-head-->", head.String(), 1)
	html = strings.Replace(html, "<!--seo-body-->", body.String(), 1)
	return c.HTML(http.StatusOK, html)
}

func addPagination(data *pageContent, path string, offset int, total int64, pageSize int) {
	if offset > 0 {
		data.Links = append(data.Links, pageLink{path + "?offset=" + strconv.Itoa(offset-pageSize), "Previous page"})
	}
	if int64(offset+pageSize) < total {
		data.Links = append(data.Links, pageLink{path + "?offset=" + strconv.Itoa(offset+pageSize), "Next page"})
	}
}

func listPageURL(c echo.Context, offset int) string {
	params := url.Values{}
	for _, key := range []string{"search", "status", "tier", "region", "live", "country", "pro", "active", "active_days", "sort", "order"} {
		if value := c.QueryParam(key); value != "" {
			params.Set(key, value)
		}
	}
	if offset > 0 {
		params.Set("offset", strconv.Itoa(offset))
	}
	result := c.Request().URL.Path
	if len(params) > 0 {
		result += "?" + params.Encode()
	}
	return result
}

func addListPagination(data *pageContent, c echo.Context, offset int, total int64) {
	if offset > 0 {
		data.Links = append(data.Links, pageLink{listPageURL(c, offset-100), "Previous page"})
	}
	if int64(offset+100) < total {
		data.Links = append(data.Links, pageLink{listPageURL(c, offset+100), "Next page"})
	}
}

type sitemapEntry struct {
	Loc string `xml:"loc"`
}
type sitemapDocument struct {
	XMLName xml.Name
	XMLNS   string         `xml:"xmlns,attr"`
	URLs    []sitemapEntry `xml:"url,omitempty"`
	Maps    []sitemapEntry `xml:"sitemap,omitempty"`
}

func (p *pagesDelivery) sitemap(c echo.Context) error {
	doc := sitemapDocument{XMLName: xml.Name{Local: "urlset"}, XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	if c.Path() == "/sitemap.xml" {
		doc.XMLName.Local = "sitemapindex"
		doc.Maps = append(doc.Maps, sitemapEntry{p.siteURL + "/sitemaps/pages/1"})
		_, leagueTotal, err := p.leagues.GetAll(c.Request().Context(), 0, 1, model.LeagueFilter{Status: "all"})
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable)
		}
		_, teamTotal, err := p.teams.GetAll(c.Request().Context(), 0, 1, model.TeamFilter{})
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable)
		}
		_, matchTotal, err := p.matchIndex.GetSitemapMatches(c.Request().Context(), 0, 1)
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable)
		}
		_, playerTotal, err := p.players.GetSitemapPlayers(c.Request().Context(), 0, 1)
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable)
		}
		for _, kind := range []struct {
			name  string
			total int64
		}{{"leagues", leagueTotal}, {"teams", teamTotal}, {"players", playerTotal}, {"matches", matchTotal}} {
			for page := int64(1); (page-1)*1000 < kind.total; page++ {
				doc.Maps = append(doc.Maps, sitemapEntry{fmt.Sprintf("%s/sitemaps/%s/%d", p.siteURL, kind.name, page)})
			}
		}
	} else {
		page, err := strconv.Atoi(c.Param("page"))
		if err != nil || page < 1 || page > 100000 {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		switch c.Param("kind") {
		case "pages":
			if page != 1 {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			doc.URLs = []sitemapEntry{{p.siteURL + "/"}, {p.siteURL + "/dpc"}, {p.siteURL + "/team"}, {p.siteURL + "/activity"}}
		case "leagues":
			rows, _, err := p.leagues.GetAll(c.Request().Context(), (page-1)*1000, 1000, model.LeagueFilter{Status: "all"})
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable)
			}
			for _, row := range rows {
				doc.URLs = append(doc.URLs, sitemapEntry{fmt.Sprintf("%s/league/%d", p.siteURL, row.ID)})
			}
		case "matches":
			rows, _, err := p.matchIndex.GetSitemapMatches(c.Request().Context(), (page-1)*1000, 1000)
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable)
			}
			for _, row := range rows {
				if row.LeagueID > 0 && positiveID(row.MatchID) {
					doc.URLs = append(doc.URLs, sitemapEntry{fmt.Sprintf("%s/match/%d/%s", p.siteURL, row.LeagueID, row.MatchID)})
				}
			}
		case "teams":
			rows, _, err := p.teams.GetAll(c.Request().Context(), (page-1)*1000, 1000, model.TeamFilter{})
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable)
			}
			for _, row := range rows {
				doc.URLs = append(doc.URLs, sitemapEntry{fmt.Sprintf("%s/team/%d", p.siteURL, row.ID)})
			}
		case "players":
			ids, _, err := p.players.GetSitemapPlayers(c.Request().Context(), (page-1)*1000, 1000)
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable)
			}
			for _, id := range ids {
				doc.URLs = append(doc.URLs, sitemapEntry{fmt.Sprintf("%s/player/%d", p.siteURL, id)})
			}
		default:
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if len(doc.URLs) == 0 {
			return echo.NewHTTPError(http.StatusNotFound)
		}
	}
	return c.XML(http.StatusOK, doc)
}

// Only document navigation errors use HTML. API routes and asset namespaces keep
// Echo's existing JSON error handling even when a browser sends Accept: text/html.
func isHTMLPageRequest(c echo.Context) bool {
	if c.Request().Method != http.MethodGet && c.Request().Method != http.MethodHead {
		return false
	}
	if !strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		return false
	}
	requestPath := c.Request().URL.Path
	if path.Ext(requestPath) != "" {
		return false
	}
	for _, prefix := range []string{"/leagues", "/teams", "/updates", "/api", "/build", "/assets", "/src", "/sitemaps", "/dpc/standings"} {
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			return false
		}
	}
	switch c.Path() {
	case "", "/*", "/", "/team", "/dpc", "/activity", "/league/:id", "/team/:id", "/match/:leagueID/:matchID":
		return true
	default:
		return false
	}
}

var errorDocument = template.Must(template.New("error-page").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="robots" content="noindex"><title>{{.Title}} | Dota 2 Leagues</title>
<link rel="stylesheet" href="/build/index.css">
<style>body{margin:0;min-height:100vh;background:#f5f5f0;color:#282d25;font:16px/1.6 'Fira Sans',Arial,sans-serif} .error-header{background:#fff;border-bottom:1px solid #dce1d4;padding:16px max(20px,calc((100vw - 1200px)/2));display:flex;align-items:center;gap:24px;flex-wrap:wrap}.error-header a{color:#60695a;text-decoration:none}.error-header nav{display:flex;gap:6px;flex-wrap:wrap}.error-header nav a{font-size:13px;padding:10px 12px;border-radius:5px}.error-header nav a:hover{background:#f0f2eb;color:#282d25}.error-header .brand{display:flex;align-items:center}.error-header img{width:129px;max-height:44px;object-fit:contain;filter:brightness(.45)}.error-main{max-width:800px;margin:72px auto;padding:0 24px}.error-code{font-size:13px;text-transform:uppercase;letter-spacing:2px;color:#626b5b}.error-main h1{font-size:clamp(28px,5vw,42px);line-height:1.2}.error-main p{color:#626b5b}.error-actions{display:flex;flex-wrap:wrap;gap:12px;margin-top:28px}.error-actions a{padding:12px 20px;border:1px solid #d9dfd1;border-radius:5px;text-decoration:none;color:#252922}.error-actions a:first-child{background:#282d25;border-color:#282d25;color:#fff}.error-actions a:focus-visible,.error-header a:focus-visible{outline:3px solid #805f16;outline-offset:4px}@media(max-width:600px){.error-header{padding:16px;gap:14px}.error-header .brand{flex-basis:100%}.error-header nav{width:100%;justify-content:space-between}.error-main{margin:40px auto;padding:0 20px}}</style></head>
<body><header class="error-header"><a class="brand" href="/" aria-label="Dota 2 Leagues home"><img src="/logo.png" alt="Dota 2 Leagues"></a><nav aria-label="Main navigation"><a href="/">Leagues</a><a href="/team">Teams</a><a href="/activity">Updates</a><a href="/dpc">DPC Standings</a></nav></header>
<main class="error-main"><div class="error-code">{{.Status}} / {{.Label}}</div><h1>{{.Title}}</h1><p>{{.Message}}</p><div class="error-actions">{{if .Retry}}<a href="{{.Retry}}">Try again</a>{{end}}<a href="/">Browse leagues</a><a href="/team">Browse teams</a></div></main></body></html>`))

func (p *pagesDelivery) errorPage(c echo.Context, status int) error {
	data := struct {
		Status                       int
		Label, Title, Message, Retry string
	}{
		Status: status, Label: "Service unavailable", Title: "We couldn’t load this page", Message: "The data service is temporarily unavailable. Please try again in a moment.", Retry: c.Request().URL.RequestURI(),
	}
	if status == http.StatusNotFound {
		data.Label, data.Title, data.Message, data.Retry = "Page not found", "This page is off the map", "The page or record you’re looking for is unavailable. Explore leagues and teams to find your next match.", ""
	}
	var body bytes.Buffer
	if err := errorDocument.Execute(&body, data); err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Add("Vary", "Accept")
	if c.Request().Method == http.MethodHead {
		c.Response().Header().Set("Content-Type", echo.MIMETextHTMLCharsetUTF8)
		return c.NoContent(status)
	}
	return c.HTML(status, body.String())
}
