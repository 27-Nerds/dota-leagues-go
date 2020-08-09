package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"dota_league/repository"
	"log"
	"strconv"
)

// GameHandler controll struct
type GameHandler struct {
	gameRepository *repository.GameRepositoryInterface
}

// NewGameHandler create new game handler struct
func NewGameHandler(gr *repository.GameRepositoryInterface) GamesHandlerInterface {
	return &GameHandler{gr}
}

// GetLiveLeagueGames preparing response from db
func (gh *GameHandler) GetLiveLeagueGames(leagueID string, offset int, limit int) (*[]model.Game, int64, error) {

	leagueIDInt, err := strconv.Atoi(leagueID)
	if err != nil {
		return nil, 0, nil
	}

	games, totalCount, err := (*gh.gameRepository).GetForLeague(leagueIDInt, offset, limit)
	if e.IsNotFound(err) {
		return nil, totalCount, &e.Error{Code: e.ENOTFOUND, Op: "GameHandler.GetLiveLeagueGames", Err: err}
	} else if err != nil {
		log.Printf("GameHandler.GetLiveLeagueGames GetForLeague error %v", err)
		return nil, totalCount, &e.Error{Op: "GameHandler.GetLiveLeagueGames", Err: err}
	}

	return games, totalCount, nil
}
