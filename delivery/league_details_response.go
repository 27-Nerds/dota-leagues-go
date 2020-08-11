package delivery

import "dota_league/model"

type LeagueDetailsFullResponse struct {
	ID                 int    `json:"league_id"`
	Name               string `json:"name"`
	Tier               int    `json:"tier"`
	Region             int    `json:"region"`
	URL                string `json:"url"`
	Description        string `json:"description"`
	StartTimestamp     int    `json:"start_timestamp"`
	EndTimestamp       int    `json:"end_timestamp"`
	RegistrationPeriod int    `json:"registration_period"`
	ProCircuitPoints   int    `json:"pro_circuit_points"`
	Status             int    `json:"status"`
	TotalPrizePool     int    `json:"total_prize_pool"`
	IsLive             bool   `json:"is_live"`
}

func newLeagueDetailsFullResponse(leagueFromDb *model.LeagueDetails) *LeagueDetailsFullResponse {
	return &LeagueDetailsFullResponse{
		ID:                 leagueFromDb.ID,
		Name:               leagueFromDb.Name,
		Tier:               leagueFromDb.Tier,
		Region:             leagueFromDb.Region,
		URL:                leagueFromDb.URL,
		Description:        leagueFromDb.Description,
		StartTimestamp:     leagueFromDb.StartTimestamp,
		EndTimestamp:       leagueFromDb.EndTimestamp,
		RegistrationPeriod: leagueFromDb.RegistrationPeriod,
		ProCircuitPoints:   leagueFromDb.ProCircuitPoints,
		Status:             leagueFromDb.Status,
		TotalPrizePool:     leagueFromDb.TotalPrizePool,
		IsLive:             leagueFromDb.IsLive,
	}
}

func generateLeagueDetailsResponse(leagueModel *model.LeagueDetails) *LeagueDetailsFullResponse {
	return newLeagueDetailsFullResponse(leagueModel)
}
