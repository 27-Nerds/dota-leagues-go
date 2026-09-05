package delivery

import "dota_league/model"

// MatchService supplies the data required by HTTP endpoints.
type MatchService interface {
	Get(leagueID int, matchID string) (*model.MatchMinimal, error)
}

// StandingsService supplies the data required by HTTP endpoints.
type StandingsService interface {
	Get() (*model.DPCStandings, error)
}

// TeamsService supplies the data required by HTTP endpoints.
type TeamsService interface {
	GetAll(offset int, limit int) (*[]model.Team, int64, error)
	GetByID(id string) (*model.Team, error)
}

// LeaguesService supplies the data required by HTTP endpoints.
type LeaguesService interface {
	GetAllActive(offset int, limit int) (*[]model.LeagueDetails, int64, error)
	GetByID(id string) (*model.LeagueDetails, error)
	GetSeries(id string, offset int, limit int) (*[]model.LeagueSeries, int64, error)
}

// LeagueResultsService supplies the data required by HTTP endpoints.
type LeagueResultsService interface {
	Get(leagueID int) (*model.DPCLeagueResults, error)
}

// GamesService supplies the data required by HTTP endpoints.
type GamesService interface {
	GetLiveLeagueGames(leagueID string, offset int, limit int) (*[]model.Game, int64, error)
}
