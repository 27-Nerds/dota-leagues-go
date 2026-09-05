package delivery

import "dota_league/model"

// GameResponse is used to generate REST response
type GameResponse struct {
	GameID     string `json:"game_id"`
	LeagueID   int    `json:"league_id"`
	Team1Name  string `json:"team1_name"`
	Team1ID    int    `json:"team1_id"`
	Team2Name  string `json:"team2_name"`
	Team2ID    int    `json:"team2_id"`
	Spectators int    `json:"spectators"`
}

func newGameResponse(gameFromDB *model.Game) *GameResponse {
	return &GameResponse{
		GameID:     gameFromDB.ServerSteamID,
		LeagueID:   gameFromDB.LeagueID,
		Team1Name:  gameFromDB.RadiantName,
		Team1ID:    gameFromDB.RadiantTeamID,
		Team2Name:  gameFromDB.DireName,
		Team2ID:    gameFromDB.DireTeamID,
		Spectators: gameFromDB.Spectators,
	}
}

func generateGameResponse(gamesFromDB []model.Game) []*GameResponse {

	gameResponse := []*GameResponse{}

	//convert model.Game to model.GameResponse
	for _, gameFromDB := range gamesFromDB {
		gameResponse = append(gameResponse, newGameResponse(&gameFromDB))
	}

	return gameResponse
}
