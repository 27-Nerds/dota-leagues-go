package delivery

import (
	"context"
	"dota_league/model"
)

// MatchService supplies the data required by HTTP endpoints.
type MatchService interface {
	Get(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error)
}

// StandingsService supplies the data required by HTTP endpoints.
type StandingsService interface {
	Get(ctx context.Context) (*model.DPCStandings, error)
}

// TeamsService supplies the data required by HTTP endpoints.
type TeamsService interface {
	GetAll(ctx context.Context, offset int, limit int) ([]model.Team, int64, error)
	GetByID(ctx context.Context, id string) (*model.Team, error)
}

// LeaguesService supplies the data required by HTTP endpoints.
type LeaguesService interface {
	GetAllActive(ctx context.Context, offset int, limit int) ([]model.LeagueDetails, int64, error)
	GetByID(ctx context.Context, id string) (*model.LeagueDetails, error)
	GetSeries(ctx context.Context, id string, offset int, limit int) ([]model.LeagueSeries, int64, error)
}

// LeagueResultsService supplies the data required by HTTP endpoints.
type LeagueResultsService interface {
	Get(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error)
}

// GamesService supplies the data required by HTTP endpoints.
type GamesService interface {
	GetLiveLeagueGames(ctx context.Context, leagueID string, offset int, limit int) ([]model.Game, int64, error)
}
