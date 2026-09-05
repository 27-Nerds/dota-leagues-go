// Package model defines league, team, player, and match data.
package model

// LiveGames struct
type LiveGames struct {
	Games []Game `json:"games"`
}

// Game struct
type Game struct {
	LeagueID      int    `json:"league_id"`
	ServerSteamID string `json:"server_steam_id"`
	RadiantName   string `json:"radiant_name"`
	RadiantLogo   string `json:"radiant_logo"`
	RadiantTeamID int    `json:"radiant_team_id"`
	DireName      string `json:"dire_name"`
	DireLogo      string `json:"dire_logo"`
	DireTeamID    int    `json:"dire_team_id"`
	Time          int    `json:"time"`
	Spectators    int    `json:"spectators"`
	LeagueNodeID  int    `json:"league_node_id"`
	SeriesID      int    `json:"series_id"`
	MatchID       string `json:"match_id"`
	DBKey         string `json:"_key,omitempty"`
}
