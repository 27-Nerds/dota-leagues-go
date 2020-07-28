package api

import (
	"dota_league/model"
	"encoding/json"
)

// LoadLiveGames loads json from the Dota API
func LoadLiveGames() (*model.LiveGames, error) {
	body, err := doRequest("https://www.dota2.com/webapi/IDOTA2League/GetLiveGames/v001?")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	liveGamesJSON := model.LiveGames{}
	err = json.NewDecoder(body).Decode(&liveGamesJSON)
	if err != nil {
		return nil, err
	}

	return &liveGamesJSON, nil
}
