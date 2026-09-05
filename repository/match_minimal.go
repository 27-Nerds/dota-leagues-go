package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
)

// MatchMinimalRepository struct
type MatchMinimalRepository struct {
	Conn Database
}

const matchMinimalCollection = "match_minimal"

// NewMatchMinimalRepository create new repository for the minimal match data in DB
func NewMatchMinimalRepository(Conn Database) *MatchMinimalRepository {
	return &MatchMinimalRepository{Conn: Conn}
}

// Store upserts the stored minimal match document
func (r *MatchMinimalRepository) Store(mm *model.MatchMinimal) error {
	const op = "MatchMinimalRepository.Store"

	mm.DBKey = mm.MatchID
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	err := r.Conn.Insert(ctx, matchMinimalCollection, mm)
	if e.ErrorCode(err) == e.ECONFLICT {
		return r.Conn.Update(ctx, matchMinimalCollection, mm.MatchID, mm)
	} else if err != nil {
		return &e.Error{Op: op, Err: err}
	}

	return nil
}

// Get returns the stored minimal match document by match id
func (r *MatchMinimalRepository) Get(matchID string) (*model.MatchMinimal, error) {
	const op = "MatchMinimalRepository.Get"

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var mm model.MatchMinimal
	_, err := r.Conn.Query(ctx, "FOR d IN "+matchMinimalCollection+" FILTER d._key == @id RETURN d", map[string]interface{}{
		"id": matchID,
	}, &mm)
	if err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	return &mm, nil
}
