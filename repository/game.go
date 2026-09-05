package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"

	"github.com/arangodb/go-driver"
)

// GameRepository repository object
type GameRepository struct {
	Conn Database
}

// NewGameRepository creates new struct
func NewGameRepository(Conn Database) *GameRepository {
	return &GameRepository{Conn}
}

// ExistsByID - check wether record exists in the DB
func (gr *GameRepository) ExistsByID(id int64) (bool, error) {

	exists, err := existsInColByID(gr.Conn, "games", strconv.FormatInt(id, 10))
	if err != nil {
		return false, &e.Error{Op: "GameRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// StoreAll - store array of records in one batch
func (gr *GameRepository) StoreAll(games *[]model.Game) error {
	// set db keys for all elements
	for i, game := range *games {
		(*games)[i].DBKey = game.ServerSteamID
	}

	// is 2 seconds enough?
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	err := gr.Conn.InsertMany(ctx, "games", *games)
	if err != nil {
		return &e.Error{Op: "GameRepository.StoreAll", Err: err}
	}

	return nil
}

// RemoveAll remove all records from the DB
func (gr *GameRepository) RemoveAll() error {

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	err := gr.Conn.ClearCollection(ctx, "games")
	if err != nil {
		return &e.Error{Op: "GameRepository.RemoveAll", Err: err}
	}

	return nil
}

// GetAll returns all leagues wgere end_timestamp is greater than current date
func (gr *GameRepository) GetAll() (*[]model.Game, error) {
	query := "FOR d IN games RETURN d"
	games, _, err := gr.queryAll(query, nil, false)
	if err != nil {
		return nil, &e.Error{Op: "GameRepository.GetAll", Err: err}
	}

	return games, err
}

// GetForLeague will return all live games for given leagueID
func (gr *GameRepository) GetForLeague(leagueID int, offset int, limit int) (*[]model.Game, int64, error) {
	query := "FOR d IN games FILTER d.league_id == @leagueID LIMIT @offset, @limit RETURN d"
	bindVars := map[string]interface{}{
		"leagueID": leagueID,
		"offset":   offset,
		"limit":    limit,
	}

	games, totalCount, err := gr.queryAll(query, bindVars, true)
	if err != nil {
		return nil, totalCount, &e.Error{Op: "GameRepository.GetForLeague", Err: err}
	}

	// if games list is empty return not found error
	if totalCount == 0 {
		return nil, totalCount, &e.Error{Code: e.ENOTFOUND, Op: "GameRepository.GetForLeague", Err: err}
	}

	return games, totalCount, err
}

// queryAll performs given query and returs array of serialized objects
func (gr *GameRepository) queryAll(query string, bindVars map[string]interface{}, withTotalCount bool) (*[]model.Game, int64, error) {
	var games []model.Game
	var totalCount int64

	ct := context.Background()
	if withTotalCount {
		ct = driver.WithQueryFullCount(context.Background(), true)
	}
	ctx, cancel := context.WithTimeout(ct, dbTimeout)
	defer cancel()

	cursor, err := gr.Conn.QueryAll(ctx, query, bindVars)
	if driver.IsNotFound(err) {
		return &games, totalCount, nil
	} else if err != nil {
		return nil, totalCount, &e.Error{Op: "GameRepository.queryAll", Err: err}
	}
	defer db.CloseCursor(cursor)

	for {
		var doc model.Game
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, totalCount, &e.Error{Op: "GameRepository.queryAll", Err: err}
		}
		games = append(games, doc)
	}
	if withTotalCount {
		totalCount = cursor.Statistics().FullCount()
	}
	return &games, totalCount, nil
}
