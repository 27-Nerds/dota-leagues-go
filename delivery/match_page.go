package delivery

import (
	"dota_league/model"
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// hero_names.json is extracted from the frontend's src/lib/heroes.js catalog.
//
//go:embed hero_names.json
var heroCatalog embed.FS
var heroNames = func() map[string]string {
	names := map[string]string{}
	data, err := heroCatalog.ReadFile("hero_names.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &names); err != nil {
		panic(err)
	}
	return names
}()

func populateMatchPage(data *pageContent, match *model.MatchMinimal, leagueID string) {
	data.Entity = matchSchema(match, data.Canonical)
	radiant, dire := strings.TrimSpace(match.Tourney.RadiantTeamName), strings.TrimSpace(match.Tourney.DireTeamName)
	if radiant == "" {
		radiant = "Radiant team"
	}
	if dire == "" {
		dire = "Dire team"
	}
	result := "Result unavailable."
	switch match.MatchOutcome {
	case 2:
		result = radiant + " won."
	case 3:
		result = dire + " won."
	case 5:
		result = "No winner."
	case 64, 65, 66, 67, 68, 69:
		result = "Match not scored."
	}
	data.Title = fmt.Sprintf("%s vs %s - Match %s | Dota 2 Leagues", radiant, dire, match.MatchID)
	data.Description = fmt.Sprintf("%s vs %s. %s Kills: %d to %d. Duration: %dm %ds.", radiant, dire, result, match.RadiantScore, match.DireScore, match.Duration/60, match.Duration%60)
	if match.StartTime > 0 {
		data.Paragraphs = append(data.Paragraphs, "Played "+time.Unix(match.StartTime, 0).UTC().Format("2 Jan 2006, 15:04")+" UTC.")
	}
	data.Links = append(data.Links, pageLink{"/league/" + leagueID, "League #" + leagueID})
	for _, team := range []struct {
		id   int
		name string
	}{{match.Tourney.RadiantTeamID, radiant}, {match.Tourney.DireTeamID, dire}} {
		if team.id > 0 {
			data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", team.id), team.name})
		}
	}
	for _, player := range match.Players {
		name := strings.TrimSpace(player.ProName)
		if name == "" {
			if player.AccountID > 0 {
				name = fmt.Sprintf("Player #%d", player.AccountID)
			} else {
				name = "Anonymous player"
			}
		}
		hero := heroNames[strconv.Itoa(player.HeroID)]
		if hero == "" {
			hero = "Unknown hero"
		}
		side := "Radiant"
		if player.TeamNumber == 1 {
			side = "Dire"
		}
		data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("%s (%s, %s): %d kills, %d deaths, %d assists. Level %d.", name, hero, side, player.Kills, player.Deaths, player.Assists, player.Level))
	}
}
