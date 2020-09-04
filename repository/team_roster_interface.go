package repository

import "dota_league/model"

// TeamRosterRepositoryInterface interface for team roster
type TeamRosterRepositoryInterface interface {
	Store(*model.TeamRoster) error
	ExistsByTeamID(TeamID int) (bool, error)
}
