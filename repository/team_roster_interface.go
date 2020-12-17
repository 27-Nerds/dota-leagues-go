package repository

import "dota_league/model"

// TeamRosterRepositoryInterface interface for team roster
type TeamRosterRepositoryInterface interface {
	Store(*model.TeamRoster) error
	ExistsByTeamID(TeamID int) (bool, error)
	GetAll(offset int, limit int) (*[]model.TeamRoster, int64, error)
	GetByID(id int) (*model.TeamRoster, error)
}
