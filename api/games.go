package api

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
	"io"
)

// LoadLiveGames loads json from the Dota API
func LoadLiveGames(ctx context.Context) (*model.LiveGames, error) {
	op := "api.LoadLiveGames"
	body, err := doRequest(ctx, "https://www.dota2.com/webapi/IDOTA2League/GetLiveGames/v001?")
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer closeResponse(body)
	liveGamesJSON := model.LiveGames{}
	err = json.NewDecoder(body).Decode(&liveGamesJSON)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &liveGamesJSON, nil
}

// GetLiveGameStats - get live game data based on serverSteamID
func GetLiveGameStats(ctx context.Context, serverSteamID string) (*model.LiveGameDetails, error) {
	op := "api.GetLiveGameStats"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2MatchStats/GetRealtimeStats/v001?server_steam_id=%s", serverSteamID)

	body, err := doRequest(ctx, url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	defer closeResponse(body)
	responseData, err := io.ReadAll(body)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

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
