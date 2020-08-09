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
func (lh *LeaguesHandler) GetAllActive(offset int, limit int) (*[]model.LeagueDetails, int64, error) {
	leaguesFromDb, totalCount, err := (*lh.LeagueDetailsRepository).GetAllActive(offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "LeaguesHandler.GetAllActive", Err: err}
	}

	return leaguesFromDb, totalCount, nil
}
