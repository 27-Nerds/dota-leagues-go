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

// GetAllActive performs DB query and return results,
// second returning value is total count
func (lh *LeaguesHandler) GetAllActive(offset int, limit int) (*[]model.LeagueDetailsResponse, int64, error) {

	leagueDetailsResponse := []model.LeagueDetailsResponse{}

	leaguesFromDb, totalCount, err := (*lh.LeagueDetailsRepository).GetAllActive(offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "LeaguesHandler.GetAllActive", Err: err}
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

	return &leagueDetailsResponse, totalCount, nil
}
