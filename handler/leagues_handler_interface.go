package handler

import "dota_league/model"

// LeaguesHandlerInterface interface for leagues handler
type LeaguesHandlerInterface interface {
	GetAllActive() (*[]model.LeagueDetailsResponse, error)
}
