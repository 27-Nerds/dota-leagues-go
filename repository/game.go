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
	query := "RETURN LENGTH(FOR d IN games FILTER d._key == @id LIMIT 1 RETURN true) > 0"
	bindVars := map[string]interface{}{
		"id": strconv.FormatInt(id, 10),
	}

	var exists bool

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := (*gr.Conn).Query(ctx, query, bindVars, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
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
	var games []model.Game

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cursor, err := (*gr.Conn).QueryAll(ctx, query, nil)
	if driver.IsNotFound(err) {
		return &games, nil
	} else if err != nil {
		return nil, &e.Error{Op: "GameRepository.GetAll", Err: err}
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
