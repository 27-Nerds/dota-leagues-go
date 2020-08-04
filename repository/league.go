package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"time"

	"github.com/arangodb/go-driver"
)

// LeagueRepository repository object
type LeagueRepository struct {
	Conn *db.Interface
}

// NewLeagueRepository creates new struct
func NewLeagueRepository(Conn *db.Interface) LeagueRepositoryInterface {
	return &LeagueRepository{Conn}
}

// Store store league model in db
func (lr *LeagueRepository) Store(l *model.League) error {
	l.DbKey = strconv.Itoa(l.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*lr.Conn).Insert(ctx, "leagues", l)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "LeagueRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "LeagueRepository.Store", Err: err}
	}

	return nil
}

// StoreAll - store array of records in one batch
func (lr *LeagueRepository) StoreAll(leagues *[]model.League) error {
	// set db keys for all elements
	for i, league := range *leagues {
		(*leagues)[i].DbKey = strconv.Itoa(league.ID)
	}

	// is 2 seconds enough?
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*lr.Conn).InsertMany(ctx, "leagues", *leagues)
	if err != nil {
		return &e.Error{Op: "LeagueRepository.StoreAll", Err: err}
	}

	return nil
}

// ExistsByID - check wether record exists in the DB
func (lr *LeagueRepository) ExistsByID(id int) (bool, error) {

	exists, err := existsInColByID(lr.Conn, "leagues", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "LeagueRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// GetByDateRange - returns array of Leagues from the db. StartDate and EndDate are timestamps
func (lr *LeagueRepository) GetByDateRange(startDate int64, endDate int64) (*[]model.League, error) {
	query := "FOR d IN leagues FILTER d.most_recent_activity >= @startDate && d.most_recent_activity <= @endDate RETURN d"
	bindVars := map[string]interface{}{
		"startDate": startDate,
		"endDate":   endDate,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cursor, err := (*lr.Conn).QueryAll(ctx, query, bindVars)
	if err != nil {
		return nil, &e.Error{Op: "LeagueRepository.GetByDateRange", Err: err}
	}

	defer cursor.Close()
	var leagues []model.League

	for {
		var doc model.League
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			// handle other errors
		}
		leagues = append(leagues, doc)
	}

	return &leagues, nil
}

// GetFromYearStart - get all tourneys in the current year
func (lr *LeagueRepository) GetFromYearStart() (*[]model.League, error) {
	now := time.Now()
	currentYear, _, _ := now.Date()
	firstOfYear := time.Date(currentYear, 1, 1, 0, 0, 0, 0, now.Location())

	leagues, err := lr.GetByDateRange(firstOfYear.Unix(), now.Unix())
	if err != nil {
		return nil, &e.Error{Op: "LeagueRepository.GetFromYearStart", Err: err}
	}

	return leagues, nil
}

// HasAnyRecord return true if there are at least one record in the DB
func (lr *LeagueRepository) HasAnyRecord() (bool, error) {
	query := "RETURN LENGTH(FOR d IN leagues LIMIT 1 RETURN true) > 0"
	var exists bool

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := (*lr.Conn).Query(ctx, query, nil, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
		return false, &e.Error{Op: "LeagueRepository.HasAnyRecords", Err: err}
	}

	return exists, nil
}

// GetAllActive returns all leagues where end_timestamp is greater than current date
func (lr *LeagueRepository) GetAllActive() (*[]model.LeagueDetails, error) {

	query := "FOR d IN leagues FILTER d.end_timestamp >= @today SORT d.tier DESC RETURN d"
	bindVars := map[string]interface{}{
		"today": time.Now().Unix(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cursor, err := (*lr.Conn).QueryAll(ctx, query, bindVars)
	if err != nil {
		return nil, &e.Error{Op: "LeagueRepository.GetAllActive", Err: err}
	}

	defer cursor.Close()
	var leagues []model.LeagueDetails

	for {
		var doc model.LeagueDetails
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		leagues = append(leagues, doc)
	}

	return &leagues, nil
}
