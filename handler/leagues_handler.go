package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"dota_league/repository"
	"strconv"
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

func (lh *LeaguesHandler) Get(id string) (*model.LeagueDetails, error) {
	leagueResponse := model.LeagueDetails{}

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return &leagueResponse, nil
	}

	data, err := (*lh.LeagueDetailsRepository).Get(idInt)
	if err != nil {
		return nil, err
	}

	return data, nil
}
