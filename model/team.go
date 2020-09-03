package model

// Team - Team api and db struct
type Team struct {
	Members []struct {
		AccountID  int  `json:"account_id"`
		TimeJoined int  `json:"time_joined"`
		Admin      bool `json:"admin"`
	} `json:"members"`
	ID                         int    `json:"team_id"`
	Name                       string `json:"name"`
	Tag                        string `json:"tag"`
	TimeCreated                int    `json:"time_created"`
	Pro                        bool   `json:"pro"`
	CountryCode                string `json:"country_code"`
	URL                        string `json:"url"`
	Wins                       int    `json:"wins"`
	Losses                     int    `json:"losses"`
	GamesPlayedTotal           int    `json:"games_played_total"`
	GamesPlayedMatchmaking     int    `json:"games_played_matchmaking"`
	RegisteredMemberAccountIds []int  `json:"registered_member_account_ids"`
	Region                     int    `json:"region"`
	URLLogo                    string `json:"url_logo"`
	DbKey                      string `json:"_key,omitempty"`
}
