package model

// Team - Team api and db struct
type Team struct {
	Members []struct {
		AccountID  int    `json:"account_id"`
		TimeJoined int    `json:"time_joined"`
		Admin      bool   `json:"admin"`
		ProName    string `json:"pro_name"`
		Role       int    `json:"role"`
	} `json:"members"`
	ID                         int          `json:"team_id"`
	Name                       string       `json:"name"`
	Tag                        string       `json:"tag"`
	TimeCreated                int          `json:"time_created"`
	Pro                        bool         `json:"pro"`
	CountryCode                string       `json:"country_code"`
	URL                        string       `json:"url"`
	Wins                       int          `json:"wins"`
	Losses                     int          `json:"losses"`
	GamesPlayedTotal           int          `json:"games_played_total"`
	GamesPlayedMatchmaking     int          `json:"games_played_matchmaking"`
	RegisteredMemberAccountIds []int        `json:"registered_member_account_ids"`
	Region                     int          `json:"region"`
	URLLogo                    string       `json:"url_logo"`
	UgcLogo                    string       `json:"ugc_logo"`
	UgcBaseLogo                string       `json:"ugc_base_logo"`
	UgcBannerLogo              string       `json:"ugc_banner_logo"`
	UgcSponsorLogo             string       `json:"ugc_sponsor_logo"`
	ColorPrimary               string       `json:"color_primary"`
	ColorSecondary             string       `json:"color_secondary"`
	Abbreviation               string       `json:"abbreviation"`
	TeamCaptain                int          `json:"team_captain"`
	DpcResults                 []DpcResult  `json:"dpc_results"`
	MemberStats                []MemberStat `json:"member_stats"`
	TeamStats                  TeamStats    `json:"team_stats"`
	UpdatedTimestamp           int64        `json:"updated_timestamp"`
	DBKey                      string       `json:"_key,omitempty"`
}

// DpcResult is a record of team results in a DPC league
type DpcResult struct {
	LeagueID  int `json:"league_id"`
	Standing  int `json:"standing"`
	Points    int `json:"points"`
	Earnings  int `json:"earnings"`
	Timestamp int `json:"timestamp"`
}

// MemberStat is a player performance stat for the team
type MemberStat struct {
	AccountID      int       `json:"account_id"`
	WinsWithTeam   int       `json:"wins_with_team"`
	LossesWithTeam int       `json:"losses_with_team"`
	TopHeroes      []TopHero `json:"top_heroes"`
	AvgKills       float64   `json:"avg_kills"`
	AvgDeaths      float64   `json:"avg_deaths"`
	AvgAssists     float64   `json:"avg_assists"`
}

// TopHero is a hero pick count for a player
type TopHero struct {
	HeroID int `json:"hero_id"`
	Picks  int `json:"picks"`
}

// TeamStats is an aggregated stats object returned by the api
type TeamStats struct {
	PlayedHeroes []PlayedHero `json:"played_heroes"`
	Farming      float64      `json:"farming"`
	Fighting     float64      `json:"fighting"`
	Versatility  float64      `json:"versatility"`
	AvgKills     float64      `json:"avg_kills"`
	AvgDeaths    float64      `json:"avg_deaths"`
	AvgDuration  float64      `json:"avg_duration"`
}

// PlayedHero is a per-hero stat for the team
type PlayedHero struct {
	HeroID     int     `json:"hero_id"`
	Picks      int     `json:"picks"`
	Wins       int     `json:"wins"`
	AvgKills   float64 `json:"avg_kills"`
	AvgDeaths  float64 `json:"avg_deaths"`
	AvgAssists float64 `json:"avg_assists"`
	AvgGpm     float64 `json:"avg_gpm"`
	AvgXpm     float64 `json:"avg_xpm"`
}
