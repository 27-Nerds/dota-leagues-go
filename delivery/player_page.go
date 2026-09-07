package delivery

import (
	"dota_league/model"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// playerDisplayName prefers the professional name, then the Steam name, then the account number.
func playerDisplayName(player *model.Player) string {
	if name := strings.TrimSpace(player.Name); name != "" {
		return name
	}
	if name := strings.TrimSpace(player.SteamName); name != "" {
		return name
	}
	return fmt.Sprintf("Player #%d", player.ID)
}

// steamProfileState summarises what the last Steam lookup found for display.
func steamProfileState(player *model.Player) string {
	switch {
	case player.SteamStatus == "" && player.SteamName == "":
		return "not checked"
	case player.SteamStatus == "unavailable":
		return "unavailable"
	case player.SteamPrivacy != "" && player.SteamPrivacy != "public":
		return player.SteamPrivacy
	case player.SteamName != "":
		return "public"
	}
	return "unknown"
}

func populatePlayerPage(data *pageContent, player *model.Player) {
	name := playerDisplayName(player)
	data.Title = name + " | Dota 2 Players"
	summary := name + " Dota 2 player profile"
	if player.TeamName != "" {
		summary += " playing for " + player.TeamName
	}
	data.Description = summary + ". Professional record, Steam profile, and recorded changes."
	data.Entity = schemaNode{"@type": "Person", "@id": strings.Split(data.Canonical, "?")[0] + "#entity", "name": name, "url": strings.Split(data.Canonical, "?")[0], "identifier": strconv.Itoa(player.ID)}
	if real := strings.TrimSpace(player.RealName); real != "" {
		data.Paragraphs = append(data.Paragraphs, "Real name: "+real)
	}
	if player.CountryCode != "" {
		data.Paragraphs = append(data.Paragraphs, "Country: "+strings.ToUpper(player.CountryCode))
	}
	if player.TeamID > 0 {
		team := player.TeamName
		if team == "" {
			team = fmt.Sprintf("Team #%d", player.TeamID)
		}
		data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", player.TeamID), team})
	}
	if roster := player.RosterTeam; roster != nil && roster.ID > 0 && roster.ID != player.TeamID {
		name := roster.Name
		if name == "" {
			name = fmt.Sprintf("Team #%d", roster.ID)
		}
		data.Paragraphs = append(data.Paragraphs, "Listed on the roster of "+name+".")
		data.Links = append(data.Links, pageLink{fmt.Sprintf("/team/%d", roster.ID), name})
	}
	if player.TotalEarnings > 0 {
		data.Paragraphs = append(data.Paragraphs, fmt.Sprintf("Recorded earnings: $%d", player.TotalEarnings))
	}
	state := steamProfileState(player)
	if player.SteamName != "" {
		line := "Steam name: " + player.SteamName
		if player.SteamLocation != "" {
			line += " · Location: " + player.SteamLocation
		}
		data.Paragraphs = append(data.Paragraphs, line)
	}
	switch state {
	case "private", "friendsonly":
		// Steam shows the name and avatar of private profiles; only the location needs a public sighting.
		note := "Steam profile is " + state + "; Steam still shows the name and avatar"
		if player.SteamPublicAt > 0 {
			note += ", other details were last seen public on " + time.UnixMilli(player.SteamPublicAt).UTC().Format("2 Jan 2006")
		}
		data.Paragraphs = append(data.Paragraphs, note+".")
	case "unavailable":
		note := "Steam profile is unavailable"
		if player.SteamUpdatedAt > 0 {
			note += "; showing data last collected " + time.UnixMilli(player.SteamUpdatedAt).UTC().Format("2 Jan 2006")
		}
		data.Paragraphs = append(data.Paragraphs, note+".")
	}
	for _, result := range player.Results {
		if result.LeagueID <= 0 {
			continue
		}
		label := result.LeagueName
		if label == "" {
			label = fmt.Sprintf("Tournament #%d", result.LeagueID)
		}
		data.Links = append(data.Links, pageLink{fmt.Sprintf("/league/%d", result.LeagueID), fmt.Sprintf("%s: place %d", label, result.Placement)})
	}
}
