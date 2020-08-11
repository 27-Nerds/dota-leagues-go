package model

// LiveGames struct
type LiveGames struct {
	Games []Game `json:"games"`
}

// Game struct
type Game struct {
	LeagueID      int    `json:"league_id"`
	ServerSteamID int64  `json:"server_steam_id"`
	RadiantName   string `json:"radiant_name"`
	RadiantTeamID int    `json:"radiant_team_id"`
	DireName      string `json:"dire_name"`
	DireTeamID    int    `json:"dire_team_id"`
	Time          int    `json:"time"`
	Spectators    int    `json:"spectators"`
	DbKey         string `json:"_key,omitempty"`
}

// GameResponse is used to generate REST response
type GameResponse struct {
	LeagueID   int    `json:"league_id"`
	Team1Name  string `json:"team1_name"`
	Team1ID    int    `json:"team1_id"`
	Team2Name  string `json:"team2_name"`
	Team2ID    int    `json:"team2_id"`
	Spectators int    `json:"spectators"`
}
