package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"log"
	"time"
)

// DPCHandler struct
type DPCHandler struct {
	DPCStandingsRepository StandingsStore
	load                   func() (*model.DPCStandings, error)
}

// NewDPCHandler return handler struct
func NewDPCHandler(sd StandingsStore, load func() (*model.DPCStandings, error)) *DPCHandler {
	return &DPCHandler{DPCStandingsRepository: sd, load: load}
}

// Get returns the stored dpc standings snapshot, refreshing from Valve after one hour
func (h *DPCHandler) Get() (*model.DPCStandings, error) {
	const op = "DPCHandler.Get"

	dpcFromDB, err := h.DPCStandingsRepository.Get()
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if dpcFromDB != nil && cacheFresh(dpcFromDB.UpdatedTimestamp) {
		return dpcFromDB, nil
	}

	dpcFromAPI, apiErr := h.load()
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	dpcFromAPI.UpdatedTimestamp = time.Now().Unix()

	if storeErr := h.DPCStandingsRepository.Store(dpcFromAPI); storeErr != nil {
		log.Printf("DPCHandler.Get: store standings error: %s", storeErr)
	}

	return dpcFromAPI, nil
}
