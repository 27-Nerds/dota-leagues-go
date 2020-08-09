package model

// LeagueDetailsData - data from dota league details endpoint
type LeagueDetailsData struct {
	Details   LeagueDetails `json:"info"`
	PrizePool struct {
		BasePrizePool  int `json:"base_prize_pool"`
		TotalPrizePool int `json:"total_prize_pool"`
	} `json:"prize_pool"`
	Streams []Stream `json:"streams"`
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
	DbKey              string   `json:"_key,omitempty"`
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
