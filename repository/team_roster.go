package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"fmt"
	"strconv"
	"time"

	"github.com/arangodb/go-driver"
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

// GetAll returns all teams
func (trr *TeamRosterRepository) GetAll(offset int, limit int) (*[]model.TeamRoster, int64, error) {

	query := fmt.Sprintf("FOR d IN team_rosters LIMIT %d, %d RETURN d", offset, limit)
	bindVars := map[string]interface{}{
		"today": time.Now().Unix(),
	}

	teams, totalCount, err := trr.queryAll(query, bindVars, true)
	if err != nil {
		return nil, 0, &e.Error{Op: "TeamRosterRepository.GetAllActive", Err: err}
	}

	return teams, totalCount, nil
}

// GetByID get league
func (trr *TeamRosterRepository) GetByID(id int) (*model.TeamRoster, error) {
	bindVars := map[string]interface{}{
		"id": strconv.Itoa(id),
	}

	query := "FOR d IN team_rosters FILTER d._key == @id RETURN d"

	var teamRoster model.TeamRoster

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := (*trr.Conn).Query(ctx, query, bindVars, &teamRoster)

	if err != nil {
		return nil, &e.Error{Op: "TeamRosterRepository.Get", Err: err}
	}

	return &teamRoster, nil
}

// queryAll performs given query and returs array of serialized objects
// second return parameter is total count of results, if withTotalCount is set to false, it will be 0
func (trr *TeamRosterRepository) queryAll(query string, bindVars map[string]interface{}, withTotalCount bool) (*[]model.TeamRoster, int64, error) {
	var totalCount int64

	ct := context.Background()
	if withTotalCount {
		ct = driver.WithQueryFullCount(nil, true)
	}
	ctx, cancel := context.WithTimeout(ct, 2*time.Second)
	defer cancel()
	cursor, err := (*trr.Conn).QueryAll(ctx, query, bindVars)
	if err != nil {
		return nil, totalCount, &e.Error{Op: "TeamRosterRepository.GetAllActive", Err: err}
	}

	defer cursor.Close()
	var leagues []model.TeamRoster

	for {
		var doc model.TeamRoster
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, totalCount, &e.Error{Op: "TeamRosterRepository.GetAllActive", Err: err}
		}
		leagues = append(leagues, doc)
	}
	if withTotalCount {
		totalCount = cursor.Statistics().FullCount()
	}

	return &leagues, totalCount, nil
}
