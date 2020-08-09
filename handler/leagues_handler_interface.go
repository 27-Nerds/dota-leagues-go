package handler

import "dota_league/model"

// LeaguesHandlerInterface interface for leagues handler
type LeaguesHandlerInterface interface {
	GetAllActive(offset int, limit int) (*[]model.LeagueDetailsResponse, int64, error)
}
