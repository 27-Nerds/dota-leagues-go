package model

// LeagueData is used to get data from Dota API
type LeagueData struct {
	Leagues []League `json:"infos"`
}

// League is a response object from the api and the DB model
type League struct {
	ID                 int    `json:"league_id"`
	Name               string `json:"name"`
	Tier               int    `json:"tier"`
	Region             int    `json:"region"`
	MostRecentActivity int64  `json:"most_recent_activity"`
	TotalPrizePool     int    `json:"total_prize_pool"`
	StartTimestamp     int64  `json:"start_timestamp"`
	EndTimestamp       int64  `json:"end_timestamp"`
	Status             int    `json:"status"`
	DbKey              string `json:"_key,omitempty"`
}
