package model

// DPCStandings - response of IDOTA2DPC/GetDPCStandings/v001.
// Items are raw maps: the arrays are empty between seasons and their item
// shape is only known from a populated sample.
type DPCStandings struct {
	UpdatedTimestamp       int64         `json:"updated_timestamp"`
	Results                []interface{} `json:"results"`
	Standings              []interface{} `json:"standings"`
	MajorWildcardStandings []interface{} `json:"major_wildcard_standings"`
	MajorGroupStandings    []interface{} `json:"major_group_standings"`
	MajorPlayoffStandings  []interface{} `json:"major_playoff_standings"`
	DBKey                  string        `json:"_key,omitempty"`
}
