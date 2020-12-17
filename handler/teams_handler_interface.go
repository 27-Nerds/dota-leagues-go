package handler

import "dota_league/model"

// TeamsHandlerInterface interface for leagues handler
type TeamsHandlerInterface interface {
	GetAll(offset int, limit int) (*[]model.TeamRoster, int64, error)
	GetById(id string) (*model.TeamRoster, error)
}
