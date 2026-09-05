package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"fmt"
	"strconv"
	"time"

	driver "github.com/arangodb/go-driver/v2/arangodb/shared"
)

// TeamRosterRepository repository object
type TeamRosterRepository struct {
	Conn Database
}

// NewTeamRosterRepository creates new struct
func NewTeamRosterRepository(Conn Database) *TeamRosterRepository {
	return &TeamRosterRepository{Conn}
}

// Store store team roster model in db
func (trr *TeamRosterRepository) Store(ctx context.Context, teamRoster *model.TeamRoster) error {
	teamRoster.DBKey = strconv.Itoa(teamRoster.TeamID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := trr.Conn.Insert(ctx, "team_rosters", teamRoster)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "TeamRosterRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "TeamRosterRepository.Store", Err: err}
	}

	return nil
}

// ExistsByTeamID check wether record exists in the DB for the given team
func (trr *TeamRosterRepository) ExistsByTeamID(ctx context.Context, TeamID int) (bool, error) {
	exists, err := existsInColByID(ctx, trr.Conn, "team_rosters", strconv.Itoa(TeamID))
	if err != nil {
		return false, &e.Error{Op: "TeamRosterRepository.ExistsByTeamID", Err: err}
	}

	return exists, nil
}

// Update updates an existing team roster record (partial merge)
func (trr *TeamRosterRepository) Update(ctx context.Context, teamRoster *model.TeamRoster) error {
	key := strconv.Itoa(teamRoster.TeamID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := trr.Conn.Update(ctx, "team_rosters", key, teamRoster)
	if err != nil {
		return &e.Error{Op: "TeamRosterRepository.Update", Err: err}
	}

	return nil
}

// GetAll returns all teams
func (trr *TeamRosterRepository) GetAll(ctx context.Context, offset int, limit int) ([]model.TeamRoster, int64, error) {

	query := fmt.Sprintf("FOR d IN team_rosters LIMIT %d, %d RETURN d", offset, limit)
	bindVars := map[string]any{
		"today": time.Now().Unix(),
	}

	teams, totalCount, err := trr.queryAll(ctx, query, bindVars, true)
	if err != nil {
		return nil, 0, &e.Error{Op: "TeamRosterRepository.GetAllActive", Err: err}
	}

	return teams, totalCount, nil
}

// GetByID get league
func (trr *TeamRosterRepository) GetByID(ctx context.Context, id int) (*model.TeamRoster, error) {
	bindVars := map[string]any{
		"id": strconv.Itoa(id),
	}

	query := "FOR d IN team_rosters FILTER d._key == @id RETURN d"

	var teamRoster model.TeamRoster

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := trr.Conn.Query(ctx, query, bindVars, &teamRoster)

	if err != nil {
		return nil, &e.Error{Op: "TeamRosterRepository.Get", Err: err}
	}

	return &teamRoster, nil
}

// queryAll performs given query and returs array of serialized objects
// second return parameter is total count of results, if withTotalCount is set to false, it will be 0
func (trr *TeamRosterRepository) queryAll(ctx context.Context, query string, bindVars map[string]any, withTotalCount bool) ([]model.TeamRoster, int64, error) {
	var totalCount int64

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	cursor, err := trr.Conn.QueryAll(ctx, query, bindVars, withTotalCount)
	if err != nil {
		return nil, totalCount, &e.Error{Op: "TeamRosterRepository.GetAllActive", Err: err}
	}

	defer db.CloseCursor(cursor)
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
		totalCount = int64(cursor.Statistics().FullCountInt)
	}

	return leagues, totalCount, nil
}
