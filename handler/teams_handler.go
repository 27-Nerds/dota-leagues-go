package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
)

type TeamsHandler struct {
	TeamRepository TeamReader
}

// NewTeamsHandler creates new TeamsHandler
func NewTeamsHandler(trr TeamReader) *TeamsHandler {
	return &TeamsHandler{
		TeamRepository: trr,
	}
}

func (th *TeamsHandler) GetAll(ctx context.Context, offset int, limit int) ([]model.Team, int64, error) {

	teamsFromDB, totalCount, err := th.TeamRepository.GetAll(ctx, offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "TeamsHandler.GetAll", Err: err}
	}

	return teamsFromDB, totalCount, nil

}

func (th *TeamsHandler) GetByID(ctx context.Context, id string) (*model.Team, error) {

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, &e.Error{Code: e.ENOTFOUND, Op: "TeamsHandler.GetByID", Err: err}
	}

	data, err := th.TeamRepository.GetByID(ctx, idInt)
	if err != nil {
		return nil, &e.Error{Op: "TeamsHandler.GetByID", Err: err}
	}

	return data, nil
}
