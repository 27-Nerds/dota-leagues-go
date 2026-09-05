package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"log"
	"time"
)

// DPCResultsHandler struct
type DPCResultsHandler struct {
	DPCResultsRepository LeagueResultsStore
	load                 func(leagueID int) (*model.DPCLeagueResults, error)
}

// NewDPCResultsHandler return handler struct
func NewDPCResultsHandler(rr LeagueResultsStore, load func(leagueID int) (*model.DPCLeagueResults, error)) *DPCResultsHandler {
	return &DPCResultsHandler{DPCResultsRepository: rr, load: load}
}

// Get returns the stored dpc results for a league, refreshing from Valve after one hour
func (h *DPCResultsHandler) Get(leagueID int) (*model.DPCLeagueResults, error) {
	const op = "DPCResultsHandler.Get"

	resultsFromDB, err := h.DPCResultsRepository.Get(leagueID)
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if resultsFromDB != nil && cacheFresh(resultsFromDB.UpdatedTimestamp) {
		return resultsFromDB, nil
	}

	resultsFromAPI, apiErr := h.load(leagueID)
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	resultsFromAPI.UpdatedTimestamp = time.Now().Unix()

	if storeErr := h.DPCResultsRepository.Store(leagueID, resultsFromAPI); storeErr != nil {
		log.Printf("DPCResultsHandler.Get: store league %d error: %s", leagueID, storeErr)
	}

	return resultsFromAPI, nil
}
