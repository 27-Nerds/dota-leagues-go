package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"time"

	driver "github.com/arangodb/go-driver/v2/arangodb/shared"
)

// LeagueRepository repository object
type LeagueRepository struct {
	Conn Database
}

// NewLeagueRepository creates new struct
func NewLeagueRepository(Conn Database) *LeagueRepository {
	return &LeagueRepository{Conn}
}

// Store store league model in db
func (lr *LeagueRepository) Store(ctx context.Context, l *model.League) error {
	l.DBKey = strconv.Itoa(l.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := lr.Conn.Insert(ctx, "leagues", l)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "LeagueRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "LeagueRepository.Store", Err: err}
	}

	return nil
}

// StoreAll - store array of records in one batch
func (lr *LeagueRepository) StoreAll(ctx context.Context, leagues []model.League) error {
	// set db keys for all elements
	for i, league := range leagues {
		leagues[i].DBKey = strconv.Itoa(league.ID)
	}

	// is 2 seconds enough?
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := lr.Conn.InsertMany(ctx, "leagues", leagues)
	if err != nil {
		return &e.Error{Op: "LeagueRepository.StoreAll", Err: err}
	}

	return nil
}

// ExistsByID - check wether record exists in the DB
func (lr *LeagueRepository) ExistsByID(ctx context.Context, id int) (bool, error) {

	exists, err := existsInColByID(ctx, lr.Conn, "leagues", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "LeagueRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// GetByDateRange - returns array of Leagues from the db. StartDate and EndDate are timestamps
func (lr *LeagueRepository) GetByDateRange(ctx context.Context, startDate int64, endDate int64) ([]model.League, error) {
	query := "FOR d IN leagues FILTER d.most_recent_activity >= @startDate && d.most_recent_activity <= @endDate RETURN d"
	bindVars := map[string]any{
		"startDate": startDate,
		"endDate":   endDate,
	}

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	cursor, err := lr.Conn.QueryAll(ctx, query, bindVars, false)
	if err != nil {
		return nil, &e.Error{Op: "LeagueRepository.GetByDateRange", Err: err}
	}

	defer db.CloseCursor(cursor)
	var leagues []model.League

	for {
		var doc model.League
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			// handle other errors
			return nil, &e.Error{Op: "LeagueRepository.GetByDateRange", Err: err}
		}
		leagues = append(leagues, doc)
	}

	return leagues, nil
}

// GetFromYearStart - get all tourneys in the current year
func (lr *LeagueRepository) GetFromYearStart(ctx context.Context) ([]model.League, error) {
	now := time.Now()
	currentYear, _, _ := now.Date()
	firstOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())

	leagues, err := lr.GetByDateRange(ctx, firstOfYear.Unix(), now.Unix())
	if err != nil {
		return nil, &e.Error{Op: "LeagueRepository.GetFromYearStart", Err: err}
	}

	return leagues, nil
}

// HasAnyRecord return true if there are at least one record in the DB
func (lr *LeagueRepository) HasAnyRecord(ctx context.Context) (bool, error) {
	query := "RETURN LENGTH(FOR d IN leagues LIMIT 1 RETURN true) > 0"
	var exists bool

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := lr.Conn.Query(ctx, query, nil, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
		return false, &e.Error{Op: "LeagueRepository.HasAnyRecords", Err: err}
	}

	return exists, nil
}

// GetAllActive returns all leagues where end_timestamp is greater than current date
func (lr *LeagueRepository) GetAllActive(ctx context.Context) ([]model.LeagueDetails, error) {

	query := "FOR d IN leagues FILTER d.end_timestamp >= @today SORT d.tier DESC RETURN d"
	bindVars := map[string]any{
		"today": time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	cursor, err := lr.Conn.QueryAll(ctx, query, bindVars, false)
	if err != nil {
		return nil, &e.Error{Op: "LeagueRepository.GetAllActive", Err: err}
	}

	defer db.CloseCursor(cursor)
	var leagues []model.LeagueDetails

	for {
		var doc model.LeagueDetails
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &e.Error{Op: "LeagueRepository.GetAllActive", Err: err}
		}
		leagues = append(leagues, doc)
	}

	return leagues, nil
}
