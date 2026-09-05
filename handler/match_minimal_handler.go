package handler

import (
	e "dota_league/error"
	"dota_league/model"
	"log"
)

// MatchMinimalHandler struct
type MatchMinimalHandler struct {
	MatchMinimalRepository MatchStore
	load                   func(leagueID int, matchID string) (*model.MatchMinimal, error)
}

// NewMatchMinimalHandler return handler struct
func NewMatchMinimalHandler(mm MatchStore, load func(leagueID int, matchID string) (*model.MatchMinimal, error)) *MatchMinimalHandler {
	return &MatchMinimalHandler{MatchMinimalRepository: mm, load: load}
}

// Get returns the stored minimal match data, fetching from Valve when not cached yet
func (h *MatchMinimalHandler) Get(leagueID int, matchID string) (*model.MatchMinimal, error) {
	const op = "MatchMinimalHandler.Get"

	matchFromDB, err := h.MatchMinimalRepository.Get(matchID)
	if !e.IsNotFound(err) && err != nil {
		return nil, &e.Error{Op: op, Err: err}
	}

	if matchFromDB != nil {
		return matchFromDB, nil
	}

	matchFromAPI, apiErr := h.load(leagueID, matchID)
	if apiErr != nil {
		return nil, &e.Error{Op: op, Err: apiErr}
	}

	if storeErr := h.MatchMinimalRepository.Store(matchFromAPI); storeErr != nil {
		log.Printf("MatchMinimalHandler.Get: store %d/%s error: %s", leagueID, matchID, storeErr)
	}

	return matchFromAPI, nil
}
