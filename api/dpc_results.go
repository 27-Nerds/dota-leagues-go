package api

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
)

// LoadDPCLeagueResults loads the DPC results panel for a league from the Dota API.
// May return an empty envelope for leagues that never published points.
func LoadDPCLeagueResults(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error) {
	const op = "api.LoadDPCLeagueResults"

	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2DPC/GetLeagueResults/v001?league_id=%d", leagueID)
	body, err := doInteractiveRequest(ctx, url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer closeResponse(body)

	var results model.DPCLeagueResults
	if err = json.NewDecoder(body).Decode(&results); err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &results, nil
}

// LoadMatchMinimal loads the minimal match data for a league match from the Dota API.
func LoadMatchMinimal(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
	const op = "api.LoadMatchMinimal"

	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2DPC/GetLeagueMatchMinimal/v001?league_id=%d&match_id=%s", leagueID, matchID)
	body, err := doInteractiveRequest(ctx, url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer closeResponse(body)

	var mm model.MatchMinimal
	if err = json.NewDecoder(body).Decode(&mm); err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &mm, nil
}
