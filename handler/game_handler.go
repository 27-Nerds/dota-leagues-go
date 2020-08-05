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
func (gh *GameHandler) GetLiveLeagueGames(leagueID string) (*[]model.GameResponse, error) {
	gameResponse := []model.GameResponse{}

	leagueIDInt, err := strconv.Atoi(leagueID)
	if err != nil {
		return &gameResponse, nil
	}

	gamesFromDb, err := (*gh.gameRepository).GetForLeague(leagueIDInt)
	if e.IsNotFound(err) {
		return nil, &e.Error{Code: e.ENOTFOUND, Op: "GameHandler.GetLiveLeagueGames", Err: err}
	} else if err != nil {
		log.Printf("GameHandler.GetLiveLeagueGames GetForLeague error %v", err)
		return nil, &e.Error{Op: "GameHandler.GetLiveLeagueGames", Err: err}
	}

	//convert model.Game to model.GameResponse
	for _, gameFromDb := range *gamesFromDb {
		g := model.GameResponse{
			LeagueID:   gameFromDb.LeagueID,
			Team1Name:  gameFromDb.RadiantName,
			Team1ID:    gameFromDb.RadiantTeamID,
			Team2Name:  gameFromDb.DireName,
			Team2ID:    gameFromDb.DireTeamID,
			Spectators: gameFromDb.Spectators,
		}
		gameResponse = append(gameResponse, g)
	}

	return &gameResponse, nil
}
