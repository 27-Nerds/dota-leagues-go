package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"dota_league/repository"
)

// LeaguesHandler struct
type LeaguesHandler struct {
	LeagueDetailsRepository *repository.LeagueDetailsRepositoryInterface
}

// NewLeaguesHandler return handler struct
func NewLeaguesHandler(ldr *repository.LeagueDetailsRepositoryInterface) LeaguesHandlerInterface {
	return &LeaguesHandler{ldr}
}

// GetAllActive performs DB query and return results
func (lh *LeaguesHandler) GetAllActive() (*[]model.LeagueDetailsResponse, error) {

	leagueDetailsResponse := []model.LeagueDetailsResponse{}

	leaguesFromDb, err := (*lh.LeagueDetailsRepository).GetAllActive()
	if err != nil {
		return nil, &e.Error{Op: "LeaguesHandler.GetAllActive", Err: err}
	}

	//convert model.Game to model.GameResponse
	for _, leagueFromDb := range *leaguesFromDb {
		l := model.LeagueDetailsResponse{
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
		leagueDetailsResponse = append(leagueDetailsResponse, l)
	}

	return &leagueDetailsResponse, nil
}
