package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
)

// PlayerReader retrieves player profiles for application requests.
type PlayerReader interface {
	GetByID(ctx context.Context, id int) (*model.Player, error)
	GetSitemapPlayers(ctx context.Context, offset, limit int) ([]int, int64, error)
}

type PlayersHandler struct{ repository PlayerReader }

func NewPlayersHandler(repository PlayerReader) *PlayersHandler {
	return &PlayersHandler{repository: repository}
}

// GetByID returns a public view of the profile without refresh bookkeeping.
func (h *PlayersHandler) GetByID(ctx context.Context, id string) (*model.Player, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil || idInt <= 0 {
		return nil, &e.Error{Code: e.ENOTFOUND, Op: "PlayersHandler.GetByID", Err: err}
	}
	player, err := h.repository.GetByID(ctx, idInt)
	if err != nil {
		return nil, &e.Error{Op: "PlayersHandler.GetByID", Err: err}
	}
	player.DBKey = ""
	player.SteamNextRefreshAt = 0
	player.ProfileCheckedAt = 0
	return player, nil
}

func (h *PlayersHandler) GetSitemapPlayers(ctx context.Context, offset, limit int) ([]int, int64, error) {
	ids, total, err := h.repository.GetSitemapPlayers(ctx, offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "PlayersHandler.GetSitemapPlayers", Err: err}
	}
	return ids, total, nil
}
