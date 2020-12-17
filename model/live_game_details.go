package model

// LiveGameDetails struct
type LiveGameDetails struct {
	DbKey string `json:"_key,omitempty"`

	Match struct {
		ServerSteamID int64 `json:"server_steam_id"`
		Matchid       int64 `json:"matchid"`
		Timestamp     int   `json:"timestamp"`
		GameTime      int   `json:"game_time"`
		GameMode      int   `json:"game_mode"`
		LeagueID      int   `json:"league_id"`
		LeagueNodeID  int   `json:"league_node_id"`
		GameState     int   `json:"game_state"`
	} `json:"match"`
	Teams []struct {
		TeamNumber  int    `json:"team_number"`
		TeamID      int    `json:"team_id"`
		TeamName    string `json:"team_name"`
		TeamTag     string `json:"team_tag"`
		Score       int    `json:"score"`
		NetWorth    int    `json:"net_worth"`
		TeamLogoURL string `json:"team_logo_url"`
		Players     []struct {
			Accountid    int     `json:"accountid"`
			Playerid     int     `json:"playerid"`
			Name         string  `json:"name"`
			Team         int     `json:"team"`
			Heroid       int     `json:"heroid"`
			Level        int     `json:"level"`
			KillCount    int     `json:"kill_count"`
			DeathCount   int     `json:"death_count"`
			AssistsCount int     `json:"assists_count"`
			DeniesCount  int     `json:"denies_count"`
			LhCount      int     `json:"lh_count"`
			Gold         int     `json:"gold"`
			X            float64 `json:"x"`
			Y            float64 `json:"y"`
			NetWorth     int     `json:"net_worth"`
			Abilities    []int   `json:"abilities"`
			Items        []int   `json:"items"`
		} `json:"players"`
	} `json:"teams"`
	Buildings []struct {
		Team      int     `json:"team"`
		Heading   float64 `json:"heading"`
		Type      int     `json:"type"`
		Lane      int     `json:"lane"`
		Tier      int     `json:"tier"`
		X         float64 `json:"x"`
		Y         float64 `json:"y"`
		Destroyed bool    `json:"destroyed"`
	} `json:"buildings"`
	GraphData struct {
		GraphGold []int `json:"graph_gold"`
	} `json:"graph_data"`
	DeltaFrame bool `json:"delta_frame"`
}
