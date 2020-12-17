package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"time"
)

type LiveGameDetails struct {
	Conn *db.Interface
}

func NewLiveGameDetailsRepository(Conn *db.Interface) LiveGameDetailsInterface {

	return &LiveGameDetails{Conn}
}

func (lgd *LiveGameDetails) Store(l *model.LiveGameDetails) error {
	l.DbKey = strconv.FormatInt(l.Match.Matchid, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*lgd.Conn).Insert(ctx, "live_game_details", l)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "LiveGameDetailsRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "LiveGameDetailsRepository.Store", Err: err}
	}

	return nil
}

func (lgd *LiveGameDetails) ExistsByID(id int64) (bool, error) {

	exists, err := existsInColByID(lgd.Conn, "live_game_details", strconv.FormatInt(id, 10))
	if err != nil {
		return false, &e.Error{Op: "LiveGameDetailsRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// Update
func (lgd *LiveGameDetails) Update(lgm *model.LiveGameDetails) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*lgd.Conn).Update(ctx, "live_game_details", lgm.DbKey, lgm)
	if err != nil {
		return &e.Error{Op: "LiveGameDetailsRepository.Update", Err: err}
	}

	return nil
}
