package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
)

type LiveGameDetails struct {
	Conn Database
}

func NewLiveGameDetailsRepository(Conn Database) *LiveGameDetails {

	return &LiveGameDetails{Conn}
}

func (lgd *LiveGameDetails) Store(ctx context.Context, l *model.LiveGameDetails) error {
	l.DBKey = l.Match.Matchid
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := lgd.Conn.Insert(ctx, "live_game_details", l)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "LiveGameDetailsRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "LiveGameDetailsRepository.Store", Err: err}
	}

	return nil
}

func (lgd *LiveGameDetails) ExistsByID(ctx context.Context, id string) (bool, error) {

	exists, err := existsInColByID(ctx, lgd.Conn, "live_game_details", id)
	if err != nil {
		return false, &e.Error{Op: "LiveGameDetailsRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// Update saves changes to stored live game details.
func (lgd *LiveGameDetails) Update(ctx context.Context, lgm *model.LiveGameDetails) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := lgd.Conn.Update(ctx, "live_game_details", lgm.DBKey, lgm)
	if err != nil {
		return &e.Error{Op: "LiveGameDetailsRepository.Update", Err: err}
	}

	return nil
}
