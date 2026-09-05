package handler

import (
	"context"
	"dota_league/model"
)

// LeagueDetailsReader retrieves league details for application requests.
type LeagueDetailsReader interface {
	GetAll(ctx context.Context, offset int, limit int, filter model.LeagueFilter) ([]model.LeagueDetails, int64, error)
	GetByID(ctx context.Context, id int) (*model.LeagueDetails, error)
}

// LeagueSeriesReader retrieves a page of league series.
type LeagueSeriesReader interface {
	GetAllByLeague(ctx context.Context, leagueID int, offset int, limit int) ([]model.LeagueSeries, int64, error)
}

// GameReader retrieves live games for a league.
type GameReader interface {
	GetForLeague(ctx context.Context, leagueID int, offset int, limit int) ([]model.Game, int64, error)
}

// TeamReader retrieves teams for application requests.
type TeamReader interface {
	GetAll(ctx context.Context, offset int, limit int, filter model.TeamFilter) ([]model.Team, int64, error)
	GetByID(ctx context.Context, id int) (*model.Team, error)
}

// StandingsStore persists the cached standings snapshot.
type StandingsStore interface {
	Get(ctx context.Context) (*model.DPCStandings, error)
	Store(context.Context, *model.DPCStandings) error
}

// LeagueResultsStore persists cached results for a league.
type LeagueResultsStore interface {
	Get(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error)
	Store(ctx context.Context, leagueID int, results *model.DPCLeagueResults) error
}

// MatchStore persists cached minimal match data.
type MatchStore interface {
	Get(ctx context.Context, matchID string) (*model.MatchMinimal, error)
	Store(context.Context, *model.MatchMinimal) error
}

// PlayerNamesReader resolves stored display names in one batch.
type PlayerNamesReader interface {
	GetNames(context.Context, []int) (map[int]string, error)
}
