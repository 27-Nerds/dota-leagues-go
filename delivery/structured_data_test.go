package delivery

import (
	"dota_league/model"
	"encoding/json"
	"strings"
	"testing"
)

func TestStructuredDataInInitialHTML(t *testing.T) {
	e := pageServer(t, nil)
	for _, tc := range []struct{ path, entity string }{{"/", ""}, {"/team/2", "SportsTeam"}, {"/league/1", "SportsEvent"}, {"/match/1/123", "SportsEvent"}} {
		t.Run(tc.path, func(t *testing.T) {
			r := getPage(e, tc.path)
			_, tail, ok := strings.Cut(r.Body.String(), `<script type="application/ld+json">`)
			if !ok {
				t.Fatal("missing JSON-LD")
			}
			raw, _, ok := strings.Cut(tail, "</script>")
			if !ok {
				t.Fatal("missing script end")
			}
			var doc struct {
				Context string           `json:"@context"`
				Graph   []map[string]any `json:"@graph"`
			}
			if err := json.Unmarshal([]byte(raw), &doc); err != nil {
				t.Fatalf("invalid JSON-LD: %v: %s", err, raw)
			}
			if doc.Context != "https://schema.org" {
				t.Fatal(doc.Context)
			}
			found := false
			breadcrumb := false
			for _, node := range doc.Graph {
				if node["@type"] == tc.entity {
					found = true
				}
				if node["@type"] == "BreadcrumbList" {
					breadcrumb = true
					items := node["itemListElement"].([]any)
					for i, item := range items {
						if item.(map[string]any)["position"] != float64(i+1) {
							t.Fatal(items)
						}
					}
				}
			}
			if tc.entity != "" && (!found || !breadcrumb) {
				t.Fatalf("missing entity/breadcrumb: %s", raw)
			}
			if tc.path == "/league/1" && !strings.Contains(raw, `\u003cscript\u003e`) {
				t.Fatalf("unsafe source text encoding: %s", raw)
			}
		})
	}
}

func TestStructuredSportsDataOmitsUnknownFacts(t *testing.T) {
	node := matchSchema(&model.MatchMinimal{MatchID: "123", StartTime: 1698513226, Duration: 2597, Tourney: model.MatchTourney{RadiantTeamID: 2, RadiantTeamName: "Team A"}}, "https://example.com/match/1/123")
	if node["startDate"] != "2023-10-28T17:13:46Z" || node["duration"] != "PT2597S" {
		t.Fatal(node)
	}
	competitors := node["competitor"].([]schemaNode)
	if len(competitors) != 1 || competitors[0]["@id"] != "https://example.com/team/2#entity" {
		t.Fatal(competitors)
	}
	for _, key := range []string{"location", "offers", "eventAttendanceMode", "eventStatus", "homeTeam", "awayTeam", "endDate"} {
		if _, ok := node[key]; ok {
			t.Fatalf("invented %s", key)
		}
	}
	empty := leagueSchema(&model.LeagueDetails{ID: 1, Name: "League", EndTimestamp: 100}, "https://example.com/league/1?offset=20")
	if empty["url"] != "https://example.com/league/1" {
		t.Fatal(empty)
	}
	for _, key := range []string{"startDate", "endDate", "location", "offers"} {
		if _, ok := empty[key]; ok {
			t.Fatalf("unexpected %s", key)
		}
	}
}
