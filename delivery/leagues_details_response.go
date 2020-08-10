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

//convert model.Game to model.GameResponse
func newLeagueDetailsResponse(leagueFromDb *model.LeagueDetails) *LeagueDetailsResponse {
	return &LeagueDetailsResponse{
		ID:             leagueFromDb.ID,
		Name:           leagueFromDb.Name,
		Tier:           leagueFromDb.Tier,
		Region:         leagueFromDb.Region,
		URL:            leagueFromDb.URL,
		Description:    leagueFromDb.Description,
		StartTimestamp: leagueFromDb.StartTimestamp,
		EndTimestamp:   leagueFromDb.EndTimestamp,
		Status:         leagueFromDb.Status,
		TotalPrizePool: leagueFromDb.TotalPrizePool,
		IsLive:         leagueFromDb.IsLive,
	}
}

func generateLeaguesDetailsResponse(leaguesModels *[]model.LeagueDetails) *[]*LeagueDetailsResponse {
	leagues := []*LeagueDetailsResponse{}

	for _, leaguesModel := range *leaguesModels {
		leagues = append(leagues, newLeagueDetailsResponse(&leaguesModel))
	}

	return &leagues
}
