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
func (lh *LeaguesHandler) GetAllActive() (*[]model.LeagueDetails, error) {

	leagues, err := (*lh.LeagueDetailsRepository).GetAllActive()
	if err != nil {
		return nil, &e.Error{Op: "LeaguesHandler.GetAllActive", Err: err}
	}

	return leagues, nil
}
