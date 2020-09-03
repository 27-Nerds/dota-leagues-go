package api

import (
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"fmt"
)

// LoadTeamDetails load team info from the api
func LoadTeamDetails(teamID int) (*model.Team, error) {

	op := "api.loadTeamDetails"
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2Teams/GetSingleTeamInfo/v0001?team_id=%d", teamID)
	body, err := doRequest(url)
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer body.Close()

	teamsJSON := model.Team{}

	err = json.NewDecoder(body).Decode(&teamsJSON)

	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &teamsJSON, nil
}
