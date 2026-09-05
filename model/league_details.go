package model

// LeagueDetailsData - data from dota league details endpoint
type LeagueDetailsData struct {
	Details     LeagueDetails `json:"info"`
	SeriesInfos []SeriesInfo  `json:"series_infos"`
	NodeGroups  []NodeGroup   `json:"node_groups"`
	PrizePool   struct {
		BasePrizePool  int `json:"base_prize_pool"`
		TotalPrizePool int `json:"total_prize_pool"`
	} `json:"prize_pool"`
	Streams []Stream `json:"streams"`
}

// SeriesInfo - single series (matchup slot) from league details endpoint
type SeriesInfo struct {
	SeriesID   int      `json:"series_id"`
	SeriesType int      `json:"series_type"`
	StartTime  int64    `json:"start_time"`
	MatchIDs   []string `json:"match_ids"`
	TeamID1    int      `json:"team_id_1"`
	TeamID2    int      `json:"team_id_2"`
}

// NodeGroup - tournament bracket node group from league details endpoint
type NodeGroup struct {
	NodeGroupID          int            `json:"node_group_id"`
	ParentNodeGroupID    int            `json:"parent_node_group_id"`
	IncomingNodeGroupIDs []int          `json:"incoming_node_group_ids"`
	AdvancingNodeGroupID int            `json:"advancing_node_group_id"`
	TeamCount            int            `json:"team_count"`
	NodeGroupType        int            `json:"node_group_type"`
	Round                int            `json:"round"`
	MaxRounds            int            `json:"max_rounds"`
	Phase                int            `json:"phase"`
	IsFinalGroup         bool           `json:"is_final_group"`
	IsCompleted          bool           `json:"is_completed"`
	TeamStandings        []TeamStanding `json:"team_standings"`
}

// TeamStanding - team standing inside a node group
type TeamStanding struct {
	Standing int    `json:"standing"`
	TeamID   int    `json:"team_id"`
	TeamName string `json:"team_name"`
	TeamTag  string `json:"team_tag"`
	TeamLogo string `json:"team_logo"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Score    string `json:"score"`
	IsPro    bool   `json:"is_pro"`
}

// LeagueSeries - league series row stored in db, enriched with team names on read
type LeagueSeries struct {
	LeagueID   int      `json:"league_id"`
	SeriesID   int      `json:"series_id"`
	SeriesType int      `json:"series_type"`
	StartTime  int64    `json:"start_time"`
	MatchIDs   []string `json:"match_ids"`
	TeamID1    int      `json:"team_id_1"`
	TeamID2    int      `json:"team_id_2"`
	TeamName1  string   `json:"team_name_1,omitempty"`
	TeamTag1   string   `json:"team_tag_1,omitempty"`
	TeamName2  string   `json:"team_name_2,omitempty"`
	TeamTag2   string   `json:"team_tag_2,omitempty"`
	DBKey      string   `json:"_key,omitempty"`
}

// LeagueDetails - details about the league model
type LeagueDetails struct {
	ID                 int      `json:"league_id"`
	Name               string   `json:"name"`
	Tier               int      `json:"tier"`
	Region             int      `json:"region"`
	URL                string   `json:"url"`
	Description        string   `json:"description"`
	StartTimestamp     int      `json:"start_timestamp"`
	EndTimestamp       int      `json:"end_timestamp"`
	ProCircuitPoints   int      `json:"pro_circuit_points"`
	Status             int      `json:"status"`
	MostRecentActivity int      `json:"most_recent_activity"`
	RegistrationPeriod int      `json:"registration_period"`
	BasePrizePool      int      `json:"base_prize_pool"`
	TotalPrizePool     int      `json:"total_prize_pool"`
	IsLive             bool     `json:"is_live"`
	UpdatedTimestamp   int64    `json:"updated_timestamp"`
	DBKey              string   `json:"_key,omitempty"`
	Streams            []Stream `json:"streams,omitempty"`
}

// Stream - stream obj
type Stream struct {
	StreamID          int    `json:"stream_id"`
	Language          int    `json:"language"`
	Name              string `json:"name"`
	BroadcastProvider int    `json:"broadcast_provider"`
	StreamURL         string `json:"stream_url"`
	VodURL            string `json:"vod_url"`
}

// LeagueFilter selects the lifecycle view and optional catalogue filters.
type LeagueFilter struct {
	Sort     string
	Order    string
	Status   string
	Search   string
	Tier     *int
	Region   *int
	LiveOnly bool
}
