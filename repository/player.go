package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
)

// PlayerRepository repository object
type PlayerRepository struct {
	Conn Database
}

// NewPlayerRepository creates new struct
func NewPlayerRepository(Conn Database) *PlayerRepository {
	return &PlayerRepository{Conn}
}

// Store store player model in db
func (pr *PlayerRepository) Store(ctx context.Context, player *model.Player) error {
	player.DBKey = strconv.Itoa(player.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := pr.Conn.Insert(ctx, "players", player)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "PlayerRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "PlayerRepository.Store", Err: err}
	}

	return nil
}

// StoreAll store array of records in one batch
func (pr *PlayerRepository) StoreAll(ctx context.Context, players []model.Player) error {
	// set db keys for all elements
	for i, player := range players {
		players[i].DBKey = strconv.Itoa(player.ID)
	}

	// is 2 seconds enough?
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := pr.Conn.InsertMany(ctx, "players", players)
	if err != nil {
		return &e.Error{Op: "PlayerRepository.StoreAll", Err: err}
	}

	return nil
}

// ExistsByID check wether record exists in the DB
func (pr *PlayerRepository) ExistsByID(ctx context.Context, id int) (bool, error) {

	exists, err := existsInColByID(ctx, pr.Conn, "players", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "PlayerRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// HasAnyRecord return true if there are at least one record in the DB
func (pr *PlayerRepository) HasAnyRecord(ctx context.Context) (bool, error) {
	query := "RETURN LENGTH(FOR d IN players LIMIT 1 RETURN true) > 0"
	var exists bool

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := pr.Conn.Query(ctx, query, nil, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
		return false, &e.Error{Op: "PlayerRepository.HasAnyRecords", Err: err}
	}

	return exists, nil
}
