package delivery

import (
	"bytes"
	appError "dota_league/error"
	"dota_league/model"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

type pageLink struct{ URL, Label string }
type pageContent struct {
	Title, Description, Canonical string
	Links                         []pageLink
	Paragraphs                    []string
}

// NewPagesDelivery serves crawlable HTML using the same services as the JSON API.
// The Svelte app progressively replaces this initial content in the browser.
func NewPagesDelivery(e *echo.Echo, leagues LeaguesService, teams TeamsService, matches MatchService, standings StandingsService, indexPath, siteURL string) error {
	origin, err := url.Parse(siteURL)
	if err != nil || origin.Host == "" || (origin.Scheme != "https" && origin.Scheme != "http") || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") {
		return fmt.Errorf("SITE_URL must be an absolute http(s) origin")
	}
	p := &pagesDelivery{leagues, teams, matches, standings, indexPath, strings.TrimRight(siteURL, "/")}
	for _, route := range []string{"/", "/dpc", "/league/:id", "/team/:id", "/match/:leagueID/:matchID"} {
		e.GET(route, p.page)
		e.HEAD(route, p.page)
	}
	e.GET("/robots.txt", func(c echo.Context) error {
		return c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: "+p.siteURL+"/sitemap.xml\n")
	})
	e.GET("/sitemap.xml", p.sitemap)
	e.GET("/sitemaps/:kind/:page", p.sitemap)
	return nil
}

type pagesDelivery struct {
	leagues   LeaguesService
	teams     TeamsService
	matches   MatchService
	standings StandingsService
	indexPath string
	siteURL   string
}

var pageHead = template.Must(template.New("head").Parse(`<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<link rel="canonical" href="{{.Canonical}}">
<meta property="og:type" content="website">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:url" content="{{.Canonical}}">`))

var pageBody = template.Must(template.New("body").Parse(`<header><a href="/">Dota 2 Leagues</a> · <a href="/dpc">DPC Standings</a></header>
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
	if err != nil || n < 0 || n%100 != 0 || n > 10000000 {
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
	switch c.Path() {
	case "/":
		var rows []model.LeagueDetails
		var total int64
		rows, total, err = p.leagues.GetAllActive(c.Request().Context(), offset, 100)
		if err == nil {
			for _, l := range rows {
				data.Links = append(data.Links, pageLink{fmt.Sprintf("/league/%d", l.ID), l.Name})
			}
			if offset > 0 && len(rows) == 0 {
				return echo.NewHTTPError(http.StatusNotFound)
			}
			addPagination(&data, "/", offset, total)
		}
	case "/league/:id":
		var league *model.LeagueDetails
		league, err = p.leagues.GetByID(c.Request().Context(), c.Param("id"))
		if err == nil && league == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err == nil {
			data.Title = league.Name + " | Dota 2 Leagues"
			data.Description = league.Description
			if data.Description == "" {
				data.Description = league.Name + " schedule, live games and match results."
			}
			data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("Prize pool: $%d", league.TotalPrizePool))
			var series []model.LeagueSeries
			var total int64
			series, total, err = p.leagues.GetSeries(c.Request().Context(), c.Param("id"), offset, 100)
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
				addPagination(&data, c.Request().URL.Path, offset, total)
			}
		}
	case "/team/:id":
		var team *model.Team
		team, err = p.teams.GetByID(c.Request().Context(), c.Param("id"))
		if err == nil && team == nil {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		if err == nil {
			data.Title = team.Name + " | Dota 2 Teams"
			data.Description = fmt.Sprintf("%s roster and tournament results. %d wins and %d losses.", team.Name, team.Wins, team.Losses)
			for _, member := range team.Members {
				if member.ProName != "" {
					data.Paragraphs = append(data.Paragraphs, member.ProName)
				}
			}
			for _, result := range team.DpcResults {
				data.Links = append(data.Links, pageLink{fmt.Sprintf("/league/%d", result.LeagueID), fmt.Sprintf("League #%d: place %d", result.LeagueID, result.Standing)})
			}
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
			data.Title = fmt.Sprintf("%s vs %s – Match %s | Dota 2 Leagues", match.Tourney.RadiantTeamName, match.Tourney.DireTeamName, match.MatchID)
			data.Description = fmt.Sprintf("Match #%s: %s %d – %d %s. Duration: %d minutes.", match.MatchID, match.Tourney.RadiantTeamName, match.RadiantScore, match.DireScore, match.Tourney.DireTeamName, match.Duration/60)
			data.Links = append(data.Links, pageLink{"/league/" + c.Param("leagueID"), "League"})
			for _, player := range match.Players {
				data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("%s: %d kills, %d deaths, %d assists", player.ProName, player.Kills, player.Deaths, player.Assists))
			}
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
	if c.Path() != "/" && c.Path() != "/league/:id" {
		data.Canonical = p.siteURL + c.Request().URL.Path
	}
	shell, err := os.ReadFile(p.indexPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Frontend build unavailable")
	}
	if !bytes.Contains(shell, []byte("<!--seo-head-->")) || !bytes.Contains(shell, []byte("<!--seo-body-->")) {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Frontend build needs updating")
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

func addPagination(data *pageContent, path string, offset int, total int64) {
	if offset > 0 {
		data.Links = append(data.Links, pageLink{path + "?offset=" + strconv.Itoa(offset-100), "Previous page"})
	}
	if int64(offset+100) < total {
		data.Links = append(data.Links, pageLink{path + "?offset=" + strconv.Itoa(offset+100), "Next page"})
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
		_, leagueTotal, err := p.leagues.GetAllActive(c.Request().Context(), 0, 1)
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable)
		}
		_, teamTotal, err := p.teams.GetAll(c.Request().Context(), 0, 1)
		if err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable)
		}
		for _, kind := range []struct {
			name  string
			total int64
		}{{"leagues", leagueTotal}, {"teams", teamTotal}} {
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
			doc.URLs = []sitemapEntry{{p.siteURL + "/"}, {p.siteURL + "/dpc"}}
		case "leagues":
			rows, _, err := p.leagues.GetAllActive(c.Request().Context(), (page-1)*1000, 1000)
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable)
			}
			for _, row := range rows {
				doc.URLs = append(doc.URLs, sitemapEntry{fmt.Sprintf("%s/league/%d", p.siteURL, row.ID)})
			}
		case "teams":
			rows, _, err := p.teams.GetAll(c.Request().Context(), (page-1)*1000, 1000)
			if err != nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable)
			}
			for _, row := range rows {
				doc.URLs = append(doc.URLs, sitemapEntry{fmt.Sprintf("%s/team/%d", p.siteURL, row.ID)})
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
