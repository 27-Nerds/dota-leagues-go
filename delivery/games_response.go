package delivery

import "dota_league/model"

// GameResponse is used to generate REST response
type GameResponse struct {
	LeagueID   int    `json:"league_id"`
	Team1Name  string `json:"team1_name"`
	Team1ID    int    `json:"team1_id"`
	Team2Name  string `json:"team2_name"`
	Team2ID    int    `json:"team2_id"`
	Spectators int    `json:"spectators"`
}

func newGameResponse(gameFromDb *model.Game) *GameResponse {
	return &GameResponse{
		LeagueID:   gameFromDb.LeagueID,
		Team1Name:  gameFromDb.RadiantName,
		Team1ID:    gameFromDb.RadiantTeamID,
		Team2Name:  gameFromDb.DireName,
		Team2ID:    gameFromDb.DireTeamID,
		Spectators: gameFromDb.Spectators,
	}
}

func generateGameResponse(gamesFromDb *[]model.Game) *[]*GameResponse {

	gameResponse := []*GameResponse{}

	//convert model.Game to model.GameResponse
	for _, gameFromDb := range *gamesFromDb {
		gameResponse = append(gameResponse, newGameResponse(&gameFromDb))
	}

	return &gameResponse
}
