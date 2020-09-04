package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"time"
)

// TeamRosterRepository repository object
type TeamRosterRepository struct {
	Conn *db.Interface
}

// NewTeamRosterRepository creates new struct
func NewTeamRosterRepository(Conn *db.Interface) TeamRosterRepositoryInterface {
	return &TeamRosterRepository{Conn}
}

// Store store team roster model in db
func (trr *TeamRosterRepository) Store(teamRoster *model.TeamRoster) error {
	teamRoster.DbKey = strconv.Itoa(teamRoster.TeamID)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*trr.Conn).Insert(ctx, "team_rosters", teamRoster)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "TeamRosterRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "TeamRosterRepository.Store", Err: err}
	}

	return nil
}

// ExistsByTeamID check wether record exists in the DB for the given team
func (trr *TeamRosterRepository) ExistsByTeamID(TeamID int) (bool, error) {
	exists, err := existsInColByID(trr.Conn, "team_rosters", strconv.Itoa(TeamID))
	if err != nil {
		return false, &e.Error{Op: "TeamRosterRepository.ExistsByTeamID", Err: err}
	}

	return exists, nil
}
