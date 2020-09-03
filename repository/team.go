package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"time"
)

// TeamRepository repository object
type TeamRepository struct {
	Conn *db.Interface
}

// NewTeamRepository creates new struct
func NewTeamRepository(Conn *db.Interface) TeamRepositoryInterface {
	return &TeamRepository{Conn}
}

// Store store team model in db
func (tr *TeamRepository) Store(team *model.Team) error {
	team.DbKey = strconv.Itoa(team.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*tr.Conn).Insert(ctx, "teams", team)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "TeamRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "TeamRepository.Store", Err: err}
	}

	return nil
}

// ExistsByID check wether record exists in the DB
func (tr *TeamRepository) ExistsByID(id int) (bool, error) {

	exists, err := existsInColByID(tr.Conn, "teams", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "TeamRepository.ExistsByID", Err: err}
	}

	return exists, nil
}
