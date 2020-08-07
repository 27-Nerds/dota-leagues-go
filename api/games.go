package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
)

// LoadLiveGames loads json from the Dota API
func LoadLiveGames() (*model.LiveGames, error) {
	op := "api.LoadLiveGames"
	body, err := doRequest("https://www.dota2.com/webapi/IDOTA2League/GetLiveGames/v001?")
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer body.Close()
	liveGamesJSON := model.LiveGames{}
	err = json.NewDecoder(body).Decode(&liveGamesJSON)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &liveGamesJSON, nil
}
