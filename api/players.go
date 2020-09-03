package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
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
