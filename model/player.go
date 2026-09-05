package model

// PlayersData - player api response
type PlayersData struct {
	Players []Player `json:"player_infos"`
}

// Player - player api and db struct
type Player struct {
	SteamName          string `json:"steam_name,omitempty"`
	SteamUpdatedAt     int64  `json:"steam_updated_at,omitempty"`
	SteamNextRefreshAt int64  `json:"steam_next_refresh_at,omitempty"`
	SteamStatus        string `json:"steam_status,omitempty"`
	// SteamLocation is the self-reported profile location, not nationality.
	SteamLocation    string `json:"steam_location,omitempty"`
	ProfileSource    string `json:"profile_source,omitempty"`
	ProfileCheckedAt int64  `json:"profile_checked_at,omitempty"`
	AvatarURL        string `json:"avatar_url,omitempty"`
	ID               int    `json:"account_id"`
	Name             string `json:"name"`
	CountryCode      string `json:"country_code"`
	FantasyRole      int    `json:"fantasy_role"`
	TeamID           int    `json:"team_id"`
	TeamName         string `json:"team_name,omitempty"`
	TeamTag          string `json:"team_tag,omitempty"`
	Sponsor          string `json:"sponsor"`
	IsLocked         bool   `json:"is_locked"`
	IsPro            bool   `json:"is_pro"`
	RealName         string `json:"real_name,omitempty"`
	Birthdate        int    `json:"birthdate,omitempty"`
	TotalEarnings    int    `json:"total_earnings"`
	Results          []struct {
		LeagueID  int `json:"league_id"`
		Placement int `json:"placement"`
		Earnings  int `json:"earnings"`
	} `json:"results,omitempty"`
	TeamURLLogo string `json:"team_url_logo,omitempty"`
	DBKey       string `json:"_key,omitempty"`
}
