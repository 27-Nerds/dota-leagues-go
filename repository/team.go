package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"

	driver "github.com/arangodb/go-driver/v2/arangodb/shared"
)

// TeamRepository repository object
type TeamRepository struct {
	Conn Database
}

// NewTeamRepository creates new struct
func NewTeamRepository(Conn Database) *TeamRepository {
	return &TeamRepository{Conn}
}

// Store store team model in db
func (tr *TeamRepository) Store(ctx context.Context, team *model.Team) error {
	team.DBKey = strconv.Itoa(team.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := tr.Conn.Insert(ctx, "teams", team)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "TeamRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "TeamRepository.Store", Err: err}
	}

	return nil
}

// ExistsByID check wether record exists in the DB
func (tr *TeamRepository) ExistsByID(ctx context.Context, id int) (bool, error) {

	exists, err := existsInColByID(ctx, tr.Conn, "teams", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "TeamRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// Update updates an existing team record (partial merge)
func (tr *TeamRepository) Update(ctx context.Context, team *model.Team) error {
	team.DBKey = strconv.Itoa(team.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := tr.Conn.Update(ctx, "teams", team.DBKey, team)
	if err != nil {
		return &e.Error{Op: "TeamRepository.Update", Err: err}
	}

	return nil
}

// GetByID get team by id
func (tr *TeamRepository) GetByID(ctx context.Context, id int) (*model.Team, error) {
	bindVars := map[string]any{
		"id": strconv.Itoa(id),
	}

	query := "FOR d IN teams FILTER d._key == @id RETURN d"

	var team model.Team
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := tr.Conn.Query(ctx, query, bindVars, &team)
	if err != nil {
		return nil, &e.Error{Op: "TeamRepository.GetByID", Err: err}
	}

	return &team, nil
}

// GetAll returns teams with pagination
func (tr *TeamRepository) GetAll(ctx context.Context, offset int, limit int) ([]model.Team, int64, error) {
	query := "FOR d IN teams LIMIT @offset, @limit RETURN d"
	bindVars := map[string]any{
		"offset": offset,
		"limit":  limit,
	}

	var teams []model.Team
	var totalCount int64

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	cursor, err := tr.Conn.QueryAll(ctx, query, bindVars, true)
	if err != nil {
		return nil, 0, &e.Error{Op: "TeamRepository.GetAll", Err: err}
	}
	defer db.CloseCursor(cursor)

	for {
		var doc model.Team
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, 0, &e.Error{Op: "TeamRepository.GetAll", Err: err}
		}
		teams = append(teams, doc)
	}

	totalCount = int64(cursor.Statistics().FullCountInt)
	return teams, totalCount, nil
}
