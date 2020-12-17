package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// GetLiveGameStats - get live game data based on serverSteamID
func GetLiveGameStats(serverSteamID int64) (*model.LiveGameDetails, error) {
	op := "api.GetLiveGameStats"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2MatchStats/GetRealtimeStats/v001?server_steam_id=%d", serverSteamID)

	body, err := doRequest(url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	defer body.Close()
	responseData, err := ioutil.ReadAll(body)

	//sometimes api returns null
	if string(responseData) == "null" {
		return nil, &e.Error{Code: e.EINTERNAL, Op: "api.GetLiveGameStats - null response recieved"}
	}

	liveGamesDetailsJSON := model.LiveGameDetails{}
	err = json.Unmarshal(responseData, &liveGamesDetailsJSON)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &liveGamesDetailsJSON, nil
}
