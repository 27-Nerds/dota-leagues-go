package api

import (
	"context"
	"dota_league/model"
	"fmt"
)

// LoadTeamDetails load team info from the api
func LoadTeamDetails(ctx context.Context, teamID int) (*model.Team, error) {

	team := &model.Team{}
	url := fmt.Sprintf("https://www.dota2.com/webapi/IDOTA2Teams/GetSingleTeamInfo/v0001?team_id=%d&get_dpc_info=true", teamID)
	if err := loadSource(ctx, "team", teamID, url, team); err != nil {
		return nil, err
	}
	return team, nil
}
