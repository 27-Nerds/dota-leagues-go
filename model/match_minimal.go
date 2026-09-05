package model

// MatchMinimalPlayer - single player entry of IDOTA2DPC/GetLeagueMatchMinimal
type MatchMinimalPlayer struct {
	AccountID  int    `json:"account_id"`
	HeroID     int    `json:"hero_id"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
	Assists    int    `json:"assists"`
	Items      []int  `json:"items"`
	PlayerSlot int    `json:"player_slot"`
	ProName    string `json:"pro_name"`
	Level      int    `json:"level"`
	TeamNumber int    `json:"team_number"`
}

// MatchTourney - tournament context block of the same response
type MatchTourney struct {
	LeagueID           int    `json:"league_id"`
	RadiantTeamID      int    `json:"radiant_team_id"`
	RadiantTeamName    string `json:"radiant_team_name"`
	RadiantTeamLogo    string `json:"radiant_team_logo"`
	RadiantTeamLogoURL string `json:"radiant_team_logo_url"`
	DireTeamID         int    `json:"dire_team_id"`
	DireTeamName       string `json:"dire_team_name"`
	DireTeamLogo       string `json:"dire_team_logo"`
	DireTeamLogoURL    string `json:"dire_team_logo_url"`
}

// MatchMinimal - response of IDOTA2DPC/GetLeagueMatchMinimal/v001
type MatchMinimal struct {
	MatchID       string               `json:"match_id"`
	StartTime     int64                `json:"start_time"`
	Duration      int                  `json:"duration"`
	GameMode      int                  `json:"game_mode"`
	Players       []MatchMinimalPlayer `json:"players"`
	Tourney       MatchTourney         `json:"tourney"`
	MatchOutcome  int                  `json:"match_outcome"`
	RadiantScore  int                  `json:"radiant_score"`
	DireScore     int                  `json:"dire_score"`
	LobbyType     int                  `json:"lobby_type"`
	IsPlayerDraft bool                 `json:"is_player_draft"`
	DBKey         string               `json:"_key,omitempty"`
}
