package handler

import (
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

// GetAllActive performs DB query and return results
func (lh *LeaguesHandler) GetAllActive() (*[]model.LeagueDetails, error) {

	leagues, err := (*lh.LeagueDetailsRepository).GetAllActive()
	if err != nil {
		return nil, err
	}

	return leagues, nil
}

func (lh *LeaguesHandler) Get(id string) (*model.LeagueDetails, error) {
	leagueResponse := model.LeagueDetails{}

	idInt, err := strconv.Atoi("11625")

	if err != nil {
		return &leagueResponse, nil
	}

	data, err := (*lh.LeagueDetailsRepository).Get(idInt)
	if err != nil {
		return nil, err
	}

	return data, nil
}
