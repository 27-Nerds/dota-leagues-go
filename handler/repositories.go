package handler

import "dota_league/model"

// LeagueDetailsReader retrieves league details for application requests.
type LeagueDetailsReader interface {
	GetAllActive(offset int, limit int) (*[]model.LeagueDetails, int64, error)
	GetByID(id int) (*model.LeagueDetails, error)
}

// LeagueSeriesReader retrieves a page of league series.
type LeagueSeriesReader interface {
	GetAllByLeague(leagueID int, offset int, limit int) (*[]model.LeagueSeries, int64, error)
}

// GameReader retrieves live games for a league.
type GameReader interface {
	GetForLeague(leagueID int, offset int, limit int) (*[]model.Game, int64, error)
}

// TeamReader retrieves teams for application requests.
type TeamReader interface {
	GetAll(offset int, limit int) (*[]model.Team, int64, error)
	GetByID(id int) (*model.Team, error)
}

// StandingsStore persists the cached standings snapshot.
type StandingsStore interface {
	Get() (*model.DPCStandings, error)
	Store(*model.DPCStandings) error
}

// LeagueResultsStore persists cached results for a league.
type LeagueResultsStore interface {
	Get(leagueID int) (*model.DPCLeagueResults, error)
	Store(leagueID int, results *model.DPCLeagueResults) error
}

// MatchStore persists cached minimal match data.
type MatchStore interface {
	Get(matchID string) (*model.MatchMinimal, error)
	Store(*model.MatchMinimal) error
}
