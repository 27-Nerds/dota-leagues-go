package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
)

// DPCStandingsRepository struct
type DPCStandingsRepository struct {
	Conn Database
}

const dpcKey = "latest"

// NewDPCStandingsRepository create new repository for the dpc standings data in DB
func NewDPCStandingsRepository(Conn Database) *DPCStandingsRepository {
	return &DPCStandingsRepository{Conn: Conn}
}

// Store upserts the latest dpc standings snapshot
func (r *DPCStandingsRepository) Store(ctx context.Context, std *model.DPCStandings) error {
	const op = "DPCStandingsRepository.Store"

	std.DBKey = dpcKey
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := r.Conn.Insert(ctx, "dpc_standings", std)
	if e.ErrorCode(err) == e.ECONFLICT {
		return r.Conn.Update(ctx, "dpc_standings", dpcKey, std)
	} else if err != nil {
		return &e.Error{Op: op, Err: err}
	}

	return nil
}

// Get returns the stored dpc standings snapshot
func (r *DPCStandingsRepository) Get(ctx context.Context) (*model.DPCStandings, error) {
	const op = "DPCStandingsRepository.Get"

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	var std model.DPCStandings
	_, err := r.Conn.Query(ctx, "FOR d IN dpc_standings FILTER d._key == @id RETURN d", map[string]any{
		"id": dpcKey,
	}, &std)
	if err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	return &std, nil
}
