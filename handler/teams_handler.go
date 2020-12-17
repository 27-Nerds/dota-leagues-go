package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"dota_league/repository"
	"strconv"
)

type TeamsHandler struct {
	TeamRosterRepository *repository.TeamRosterRepositoryInterface
}

// NewTeamsHandler creates new TeamsHandler
func NewTeamsHandler(trr *repository.TeamRosterRepositoryInterface) TeamsHandlerInterface {
	return &TeamsHandler{
		TeamRosterRepository: trr,
	}
}

func (th *TeamsHandler) GetAll(offset int, limit int) (*[]model.TeamRoster, int64, error) {

	teamsFromDB, totalCount, err := (*th.TeamRosterRepository).GetAll(offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "TeamsHandler.GetAll", Err: err}
	}

	return teamsFromDB, totalCount, nil

}

func (th *TeamsHandler) GetById(id string) (*model.TeamRoster, error) {

	idInt, err := strconv.Atoi(id)

	if err != nil {
		// return &leagueResponse, nil
	}

	data, err := (*th.TeamRosterRepository).GetByID(idInt)

	if err != nil {
		return nil, err
	}

	return data, nil
}
