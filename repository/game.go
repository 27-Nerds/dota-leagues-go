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

// GameRepository repository object
type GameRepository struct {
	Conn *db.Interface
}

// NewGameRepository creates new struct
func NewGameRepository(Conn *db.Interface) GameRepositoryInterface {
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
		(*games)[i].DbKey = strconv.FormatInt(game.ServerSteamID, 10)
	}

	// is 2 seconds enough?
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*gr.Conn).InsertMany(ctx, "games", *games)
	if err != nil {
		return &e.Error{Op: "GameRepository.StoreAll", Err: err}
	}

	return nil
}

// RemoveAll remove all records from the DB
func (gr *GameRepository) RemoveAll() error {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*gr.Conn).ClearCollection(ctx, "games")
	if err != nil {
		return &e.Error{Op: "GameRepository.RemoveAll", Err: err}
	}

	return nil
}

// GetAll returns all leagues wgere end_timestamp is greater than current date
func (gr *GameRepository) GetAll() (*[]model.Game, error) {
	query := "FOR d IN games RETURN d"
	games, err := gr.queryAll(query, nil)
	if err != nil {
		return nil, &e.Error{Op: "GameRepository.GetAll", Err: err}
	}

	return games, err
}

// GetForLeague will return all live games for given leagueId
func (gr *GameRepository) GetForLeague(leagueID int) (*[]model.Game, error) {
	query := "FOR d IN games FILTER d.league_id == @leagueId  RETURN d"
	bindVars := map[string]interface{}{
		"leagueId": leagueID,
	}

	games, err := gr.queryAll(query, bindVars)
	if err != nil {
		return nil, &e.Error{Op: "GameRepository.GetForLeague", Err: err}
	}

	// if games list is empty return not found error
	if len(*games) == 0 {
		return nil, &e.Error{Code: e.ENOTFOUND, Op: "GameRepository.GetForLeague", Err: err}
	}

	return games, err
}

// queryAll performs given query and returs array of serialized objects
func (gr *GameRepository) queryAll(query string, bindVars map[string]interface{}) (*[]model.Game, error) {
	var games []model.Game

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cursor, err := (*gr.Conn).QueryAll(ctx, query, bindVars)
	if driver.IsNotFound(err) {
		return &games, nil
	} else if err != nil {
		return nil, &e.Error{Op: "GameRepository.queryAll", Err: err}
	}
	defer cursor.Close()

	for {
		var doc model.Game
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, err
		}
		games = append(games, doc)
	}

	return &games, nil
}
