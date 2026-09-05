package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
)

// MatchMinimalHandler struct
type MatchMinimalHandler struct {
	MatchMinimalRepository MatchStore
	load                   func(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error)
}

// NewMatchMinimalHandler return handler struct
func NewMatchMinimalHandler(mm MatchStore, load func(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error)) *MatchMinimalHandler {
	return &MatchMinimalHandler{MatchMinimalRepository: mm, load: load}
}

// Get returns the stored minimal match data, fetching from Valve when not cached yet
func (h *MatchMinimalHandler) Get(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
	const op = "MatchMinimalHandler.Get"

	matchFromDB, err := h.MatchMinimalRepository.Get(ctx, matchID)
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if matchFromDB != nil {
		return matchFromDB, nil
	}

	matchFromAPI, apiErr := h.load(ctx, leagueID, matchID)
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	if storeErr := h.MatchMinimalRepository.Store(ctx, matchFromAPI); storeErr != nil {
		slog.WarnContext(ctx, "cache match", "league_id", leagueID, "match_id", matchID, "error", storeErr)
	}

	return matchFromAPI, nil
}
