package handler

import "dota_league/model"

// GamesHandlerInterface interface for leagues handler
type GamesHandlerInterface interface {
	GetLiveLeagueGames(leagueID string, offset int, limit int) (*[]model.Game, int64, error)
}
