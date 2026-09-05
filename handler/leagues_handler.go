package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
)

// LeaguesHandler struct
type LeaguesHandler struct {
	LeagueDetailsRepository LeagueDetailsReader
	SeriesRepo              LeagueSeriesReader
}

// NewLeaguesHandler return handler struct
func NewLeaguesHandler(ldr LeagueDetailsReader, lsr LeagueSeriesReader) *LeaguesHandler {
	return &LeaguesHandler{ldr, lsr}
}

// GetSeries returns league series (matchup schedule) with team names joined from teams collection
func (lh *LeaguesHandler) GetSeries(ctx context.Context, id string, offset int, limit int) ([]model.LeagueSeries, int64, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, 0, &e.Error{Code: e.ENOTFOUND, Op: "GetSeries", Err: err}
	}

	seriesFromDB, totalCount, err := lh.SeriesRepo.GetAllByLeague(ctx, idInt, offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "LeaguesHandler.GetSeries", Err: err}
	}

	return seriesFromDB, totalCount, nil
}

// GetAllActive performs DB query and return results,
// second returning value is total count
func (lh *LeaguesHandler) GetAllActive(ctx context.Context, offset int, limit int) ([]model.LeagueDetails, int64, error) {
	leaguesFromDB, totalCount, err := lh.LeagueDetailsRepository.GetAllActive(ctx, offset, limit)
	if err != nil {
		return nil, 0, &e.Error{Op: "LeaguesHandler.GetAllActive", Err: err}
	}

	return leaguesFromDB, totalCount, nil
}

// GetByID performs DB query and return results
func (lh *LeaguesHandler) GetByID(ctx context.Context, id string) (*model.LeagueDetails, error) {
	leagueResponse := model.LeagueDetails{}

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return &leagueResponse, nil
	}

	data, err := lh.LeagueDetailsRepository.GetByID(ctx, idInt)

	if err != nil {
		return nil, err
	}

	return data, nil
}
