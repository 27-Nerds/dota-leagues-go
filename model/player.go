package model

// PlayersData - player api response
type PlayersData struct {
	Players []Player `json:"player_infos"`
}

// Player - player api and db struct
type Player struct {
	ID            int    `json:"account_id"`
	Name          string `json:"name"`
	CountryCode   string `json:"country_code"`
	FantasyRole   int    `json:"fantasy_role"`
	TeamID        int    `json:"team_id"`
	TeamName      string `json:"team_name,omitempty"`
	TeamTag       string `json:"team_tag,omitempty"`
	Sponsor       string `json:"sponsor"`
	IsLocked      bool   `json:"is_locked"`
	IsPro         bool   `json:"is_pro"`
	RealName      string `json:"real_name,omitempty"`
	Birthdate     int    `json:"birthdate,omitempty"`
	TotalEarnings int    `json:"total_earnings"`
	Results       []struct {
		LeagueID  int `json:"league_id"`
		Placement int `json:"placement"`
		Earnings  int `json:"earnings"`
	} `json:"results,omitempty"`
	TeamURLLogo string `json:"team_url_logo,omitempty"`
	DbKey       string `json:"_key,omitempty"`
}
