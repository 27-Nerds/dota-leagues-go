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

func newLeagueDetailsFullResponse(leagueFromDB *model.LeagueDetails) *LeagueDetailsFullResponse {
	return &LeagueDetailsFullResponse{
		ID:                 leagueFromDB.ID,
		Name:               leagueFromDB.Name,
		Tier:               leagueFromDB.Tier,
		Region:             leagueFromDB.Region,
		URL:                leagueFromDB.URL,
		Description:        leagueFromDB.Description,
		StartTimestamp:     leagueFromDB.StartTimestamp,
		EndTimestamp:       leagueFromDB.EndTimestamp,
		RegistrationPeriod: leagueFromDB.RegistrationPeriod,
		ProCircuitPoints:   leagueFromDB.ProCircuitPoints,
		Status:             leagueFromDB.Status,
		TotalPrizePool:     leagueFromDB.TotalPrizePool,
		IsLive:             leagueFromDB.IsLive,
	}
}

func generateLeagueDetailsResponse(leagueModel *model.LeagueDetails) *LeagueDetailsFullResponse {
	return newLeagueDetailsFullResponse(leagueModel)
}
