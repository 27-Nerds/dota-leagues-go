package delivery

import "dota_league/model"

// LeagueDetailsResponse is used to generate REST response
type LeagueDetailsResponse struct {
	ID             int    `json:"league_id"`
	Name           string `json:"name"`
	Tier           int    `json:"tier"`
	Region         int    `json:"region"`
	URL            string `json:"url"`
	Description    string `json:"description"`
	StartTimestamp int    `json:"start_timestamp"`
	EndTimestamp   int    `json:"end_timestamp"`
	Status         int    `json:"status"`
	TotalPrizePool int    `json:"total_prize_pool"`
	IsLive         bool   `json:"is_live"`
}

// convert model.Game to model.GameResponse
func newLeagueDetailsResponse(leagueFromDB *model.LeagueDetails) *LeagueDetailsResponse {
	return &LeagueDetailsResponse{
		ID:             leagueFromDB.ID,
		Name:           leagueFromDB.Name,
		Tier:           leagueFromDB.Tier,
		Region:         leagueFromDB.Region,
		URL:            leagueFromDB.URL,
		Description:    leagueFromDB.Description,
		StartTimestamp: leagueFromDB.StartTimestamp,
		EndTimestamp:   leagueFromDB.EndTimestamp,
		Status:         leagueFromDB.Status,
		TotalPrizePool: leagueFromDB.TotalPrizePool,
		IsLive:         leagueFromDB.IsLive,
	}
}

func generateLeaguesDetailsResponse(leaguesModels []model.LeagueDetails) []*LeagueDetailsResponse {
	leagues := []*LeagueDetailsResponse{}

	for _, leaguesModel := range leaguesModels {
		leagues = append(leagues, newLeagueDetailsResponse(&leaguesModel))
	}

	return leagues
}
