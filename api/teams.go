package api

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
)

// LoadTeamDetails load team info from the api
func LoadTeamDetails(ctx context.Context, teamID int) (*model.Team, error) {

	op := "api.loadTeamDetails"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2Teams/GetSingleTeamInfo/v0001?team_id=%d&get_dpc_info=true", teamID)
	body, err := doRequest(ctx, url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer closeResponse(body)

	teamsJSON := model.Team{}

	err = json.NewDecoder(body).Decode(&teamsJSON)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &teamsJSON, nil
}
