package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
)

// DPCResultsRepository struct
type DPCResultsRepository struct {
	Conn Database
}

const dpcResultsCollection = "league_results"

// NewDPCResultsRepository create new repository for the per-league dpc results data in DB
func NewDPCResultsRepository(Conn Database) *DPCResultsRepository {
	return &DPCResultsRepository{Conn: Conn}
}

// Store upserts the stored results snapshot for a league
func (r *DPCResultsRepository) Store(ctx context.Context, leagueID int, res *model.DPCLeagueResults) error {
	const op = "DPCResultsRepository.Store"

	key := strconv.Itoa(leagueID)
	res.DBKey = key
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.Conn.Insert(ctx, dpcResultsCollection, res)
	if e.ErrorCode(err) == e.ECONFLICT {
		return r.Conn.Update(ctx, dpcResultsCollection, key, res)
	} else if err != nil {
		return &e.Error{Op: op, Err: err}
	}

	return nil
}

// Get returns the stored results snapshot for a league
func (r *DPCResultsRepository) Get(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error) {
	const op = "DPCResultsRepository.Get"

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var res model.DPCLeagueResults
	_, err := r.Conn.Query(ctx, "FOR d IN "+dpcResultsCollection+" FILTER d._key == @id RETURN d", map[string]any{
		"id": strconv.Itoa(leagueID),
	}, &res)
	if err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	return &res, nil
}
