package worker

import "dota_league/model"

// LiveGameDetailsRepository provides persistence operations needed by background refreshes.
type LiveGameDetailsRepository interface {
	Store(*model.LiveGameDetails) error
	ExistsByID(id string) (bool, error)
	Update(lgm *model.LiveGameDetails) error
}

// PlayerRepository provides persistence operations needed by background refreshes.
type PlayerRepository interface {
	Store(*model.Player) error
	StoreAll(*[]model.Player) error
	ExistsByID(id int) (bool, error)
	HasAnyRecord() (bool, error)
}

// TeamRepository provides persistence operations needed by background refreshes.
type TeamRepository interface {
	Store(*model.Team) error
	Update(*model.Team) error
	GetByID(id int) (*model.Team, error)
	ExistsByID(id int) (bool, error)
}

// LeagueDetailsRepository provides persistence operations needed by background refreshes.
type LeagueDetailsRepository interface {
	Store(*model.LeagueDetails) error
	Update(*model.LeagueDetails) error
	ExistsByID(id int) (bool, error)
	GetAllActiveForTiers(tiers []int) (*[]model.LeagueDetails, error)
	GetByID(id int) (*model.LeagueDetails, error)
	UpdateLiveStatus(key int, newStatus bool) error
	UpdateTotalPrizePool(key int, prizePool int) error
	SetAllAsNotLive() error
}

// LeagueRepository provides persistence operations needed by background refreshes.
type LeagueRepository interface {
	Store(*model.League) error
	StoreAll(*[]model.League) error
	ExistsByID(id int) (bool, error)
	HasAnyRecord() (bool, error)
	GetAllActive() (*[]model.LeagueDetails, error)
}

// TeamRosterRepository provides persistence operations needed by background refreshes.
type TeamRosterRepository interface {
	Store(*model.TeamRoster) error
	Update(*model.TeamRoster) error
	ExistsByTeamID(TeamID int) (bool, error)
}

// GameRepository provides persistence operations needed by background refreshes.
type GameRepository interface {
	StoreAll(games *[]model.Game) error
	GetAll() (*[]model.Game, error)
	RemoveAll() error
}

// LeagueSeriesRepository provides persistence operations needed by background refreshes.
type LeagueSeriesRepository interface {
	ReplaceAllForLeague(leagueID int, series []model.SeriesInfo) error
}
