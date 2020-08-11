package handler

import "dota_league/model"

// LeaguesHandlerInterface interface for leagues handler
type LeaguesHandlerInterface interface {
	GetAllActive(offset int, limit int) (*[]model.LeagueDetails, int64, error)
	GetById(id string) (*model.LeagueDetails, error)
}
