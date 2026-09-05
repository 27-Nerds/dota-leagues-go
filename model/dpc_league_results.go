package model

// DPCResultStanding - single row of IDOTA2DPC/GetLeagueResults "results"
type DPCResultStanding struct {
	Standing         int    `json:"standing"`
	TeamID           int    `json:"team_id"`
	TeamName         string `json:"team_name"`
	TeamLogo         string `json:"team_logo"`
	TeamLogoURL      string `json:"team_logo_url"`
	Points           int    `json:"points"`
	Earnings         int    `json:"earnings"`
	Timestamp        int64  `json:"timestamp"`
	TeamAbbreviation string `json:"team_abbreviation"`
}

// DPCLeagueResults - response of IDOTA2DPC/GetLeagueResults/v001 for one league
type DPCLeagueResults struct {
	UpdatedTimestamp int64               `json:"updated_timestamp"`
	Results          []DPCResultStanding `json:"results"`
	Points           []int               `json:"points"`
	Dollars          []int               `json:"dollars"`
	DBKey            string              `json:"_key,omitempty"`
}
