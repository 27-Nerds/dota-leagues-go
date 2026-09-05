package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"time"
)

// DPCHandler struct
type DPCHandler struct {
	DPCStandingsRepository StandingsStore
	load                   func(ctx context.Context) (*model.DPCStandings, error)
}

// NewDPCHandler return handler struct
func NewDPCHandler(sd StandingsStore, load func(ctx context.Context) (*model.DPCStandings, error)) *DPCHandler {
	return &DPCHandler{DPCStandingsRepository: sd, load: load}
}

// Get returns the stored dpc standings snapshot, refreshing from Valve after one hour
func (h *DPCHandler) Get(ctx context.Context) (*model.DPCStandings, error) {
	const op = "DPCHandler.Get"

	dpcFromDB, err := h.DPCStandingsRepository.Get(ctx)
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if dpcFromDB != nil && cacheFresh(dpcFromDB.UpdatedTimestamp) {
		return dpcFromDB, nil
	}

	dpcFromAPI, apiErr := h.load(ctx)
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	dpcFromAPI.UpdatedTimestamp = time.Now().Unix()

	if storeErr := h.DPCStandingsRepository.Store(ctx, dpcFromAPI); storeErr != nil {
		slog.WarnContext(ctx, "cache standings", "error", storeErr)
	}

	return dpcFromAPI, nil
}
