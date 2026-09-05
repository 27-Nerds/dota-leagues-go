package delivery

import (
	"dota_league/model"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

type schemaNode map[string]any

func sportsEntity(kind, name, canonical, identifier string) schemaNode {
	return schemaNode{"@type": kind, "@id": strings.Split(canonical, "?")[0] + "#entity", "name": name, "url": strings.Split(canonical, "?")[0], "identifier": identifier, "sport": "Dota 2"}
}

func leagueSchema(league *model.LeagueDetails, canonical string) schemaNode {
	node := sportsEntity("SportsEvent", league.Name, canonical, fmt.Sprint(league.ID))
	if league.Description != "" {
		node["description"] = league.Description
	}
	// League pages display calendar dates, not exact start/end times.
	if league.StartTimestamp > 0 {
		node["startDate"] = time.Unix(int64(league.StartTimestamp), 0).UTC().Format("2006-01-02")
		if league.EndTimestamp >= league.StartTimestamp {
			node["endDate"] = time.Unix(int64(league.EndTimestamp), 0).UTC().Format("2006-01-02")
		}
	}
	return node
}

func matchSchema(match *model.MatchMinimal, canonical string) schemaNode {
	name := fmt.Sprintf("Match #%s", match.MatchID)
	if match.Tourney.RadiantTeamName != "" && match.Tourney.DireTeamName != "" {
		name = match.Tourney.RadiantTeamName + " vs " + match.Tourney.DireTeamName
	}
	node := sportsEntity("SportsEvent", name, canonical, match.MatchID)
	if match.StartTime > 0 {
		node["startDate"] = time.Unix(match.StartTime, 0).UTC().Format(time.RFC3339)
	}
	if match.Duration > 0 {
		node["duration"] = fmt.Sprintf("PT%dS", match.Duration)
	}
	origin := strings.Split(canonical, "/match/")[0]
	competitors := []schemaNode{}
	for _, team := range []struct {
		id   int
		name string
	}{{match.Tourney.RadiantTeamID, match.Tourney.RadiantTeamName}, {match.Tourney.DireTeamID, match.Tourney.DireTeamName}} {
		if strings.TrimSpace(team.name) == "" {
			continue
		}
		competitor := schemaNode{"@type": "SportsTeam", "name": team.name, "sport": "Dota 2"}
		if team.id > 0 {
			competitor = sportsEntity("SportsTeam", team.name, fmt.Sprintf("%s/team/%d", origin, team.id), fmt.Sprint(team.id))
		}
		competitors = append(competitors, competitor)
	}
	if len(competitors) > 0 {
		node["competitor"] = competitors
	}
	return node
}

func structuredData(data pageContent, origin, route, leagueID string) (template.JS, error) {
	websiteID := origin + "/#website"
	page := schemaNode{"@type": "WebPage", "@id": data.Canonical + "#webpage", "url": data.Canonical, "name": data.Title, "description": data.Description, "isPartOf": schemaNode{"@id": websiteID}}
	if route == "/" || route == "/team" || route == "/activity" {
		page["@type"] = "CollectionPage"
	}
	graph := []schemaNode{{"@type": "WebSite", "@id": websiteID, "url": origin + "/", "name": "Dota 2 Leagues"}, page}
	if data.Entity != nil {
		page["mainEntity"] = schemaNode{"@id": data.Entity["@id"]}
		graph = append(graph, data.Entity)
	}
	crumbs := []schemaNode{}
	add := func(name, url string) {
		crumbs = append(crumbs, schemaNode{"@type": "ListItem", "position": len(crumbs) + 1, "name": name, "item": url})
	}
	switch route {
	case "/league/:id":
		add("Leagues", origin+"/")
	case "/team/:id":
		add("Teams", origin+"/team")
	case "/match/:leagueID/:matchID":
		add("Leagues", origin+"/")
		add("League #"+leagueID, origin+"/league/"+leagueID)
	}
	if len(crumbs) > 0 {
		name := data.Title
		if data.Entity != nil {
			if entityName, ok := data.Entity["name"].(string); ok {
				name = entityName
			}
		}
		add(name, strings.Split(data.Canonical, "?")[0])
		breadcrumb := schemaNode{"@type": "BreadcrumbList", "@id": data.Canonical + "#breadcrumbs", "itemListElement": crumbs}
		page["breadcrumb"] = schemaNode{"@id": breadcrumb["@id"]}
		graph = append(graph, breadcrumb)
	}
	encoded, err := json.Marshal(schemaNode{"@context": "https://schema.org", "@graph": graph})
	if err != nil {
		return "", err
	}
	// json.Marshal escapes HTML characters, including closing script tags in source names.
	return template.JS(encoded), nil
}
