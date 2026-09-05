package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"time"
)

// DPCResultsHandler struct
type DPCResultsHandler struct {
	DPCResultsRepository LeagueResultsStore
	load                 func(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error)
}

// NewDPCResultsHandler return handler struct
func NewDPCResultsHandler(rr LeagueResultsStore, load func(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error)) *DPCResultsHandler {
	return &DPCResultsHandler{DPCResultsRepository: rr, load: load}
}

// Get returns the stored dpc results for a league, refreshing from Valve after one hour
func (h *DPCResultsHandler) Get(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error) {
	const op = "DPCResultsHandler.Get"

	resultsFromDB, err := h.DPCResultsRepository.Get(ctx, leagueID)
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if resultsFromDB != nil && cacheFresh(resultsFromDB.UpdatedTimestamp) {
		return resultsFromDB, nil
	}

	resultsFromAPI, apiErr := h.load(ctx, leagueID)
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	resultsFromAPI.UpdatedTimestamp = time.Now().Unix()

	if storeErr := h.DPCResultsRepository.Store(ctx, leagueID, resultsFromAPI); storeErr != nil {
		slog.WarnContext(ctx, "cache league results", "league_id", leagueID, "error", storeErr)
	}

	return resultsFromAPI, nil
}
