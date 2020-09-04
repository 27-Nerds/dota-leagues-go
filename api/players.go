package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
)

// LoadPlayers load players from dota api
func LoadPlayers() (*model.PlayersData, error) {
	op := "api.LoadPlayers"
	body, err := doRequest("https://www.dota2.com/webapi/IDOTA2Fantasy/GetProPlayerInfo/v001")
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer body.Close()

	playersDataJSON := model.PlayersData{}
	err = json.NewDecoder(body).Decode(&playersDataJSON)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &playersDataJSON, nil
}

// LoadSinglePlayer load info from dota api for given playerID
func LoadSinglePlayer(playerID int) (*model.Player, error) {
	op := "api.LoadSinglePlayer"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2DPC/GetPlayerInfo/v001?account_id=%d", playerID)
	body, err := doRequest(url)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer body.Close()

	playerJSON := model.Player{}

	err = json.NewDecoder(body).Decode(&playerJSON)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &playerJSON, nil
}
