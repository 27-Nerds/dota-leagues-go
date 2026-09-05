package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"slices"
	"strings"
	"time"
)

// MatchMinimalHandler struct
type MatchMinimalHandler struct {
	MatchMinimalRepository MatchStore
	players                PlayerNamesReader
	load                   func(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error)
}

// NewMatchMinimalHandler return handler struct
func NewMatchMinimalHandler(mm MatchStore, players PlayerNamesReader, load func(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error)) *MatchMinimalHandler {
	return &MatchMinimalHandler{MatchMinimalRepository: mm, players: players, load: load}
}

// Get returns the stored minimal match data, fetching from Valve when not cached yet
func (h *MatchMinimalHandler) Get(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
	const op = "MatchMinimalHandler.Get"
	// Bound the complete cache-miss path, including the API queue wait.
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	matchFromDB, err := h.MatchMinimalRepository.Get(ctx, matchID)
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if matchFromDB != nil {
		return h.withPlayerNames(ctx, matchFromDB), nil
	}

	matchFromAPI, apiErr := h.load(ctx, leagueID, matchID)
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	if storeErr := h.MatchMinimalRepository.Store(ctx, matchFromAPI); storeErr != nil {
		slog.WarnContext(ctx, "cache match", "league_id", leagueID, "match_id", matchID, "error", storeErr)
	}

	return h.withPlayerNames(ctx, matchFromAPI), nil
}

// Resolve names at read time so newly stored profiles also improve cached matches.
// Keep the original match names and cache unchanged: profile names may be newer.
func (h *MatchMinimalHandler) withPlayerNames(ctx context.Context, match *model.MatchMinimal) *model.MatchMinimal {
	if match == nil || h.players == nil {
		return match
	}
	ids := []int{}
	for _, player := range match.Players {
		if player.AccountID > 0 && strings.TrimSpace(player.ProName) == "" && !slices.Contains(ids, player.AccountID) {
			ids = append(ids, player.AccountID)
		}
	}
	if len(ids) == 0 {
		return match
	}
	names, err := h.players.GetNames(ctx, ids)
	if err != nil {
		slog.WarnContext(ctx, "resolve match player names", "match_id", match.MatchID, "error", err)
		return match
	}
	result := *match
	result.Players = slices.Clone(match.Players)
	for i := range result.Players {
		player := &result.Players[i]
		if player.AccountID > 0 && strings.TrimSpace(player.ProName) == "" {
			if name := strings.TrimSpace(names[player.AccountID]); name != "" {
				player.ProName = name
			}
		}
	}
	return &result
}
