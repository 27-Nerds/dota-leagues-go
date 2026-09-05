package worker

import (
	"context"
	"dota_league/model"
	"time"
)

// LiveGameDetailsRepository provides persistence operations needed by background refreshes.
type LiveGameDetailsRepository interface {
	Store(context.Context, *model.LiveGameDetails) error
	ExistsByID(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, lgm *model.LiveGameDetails) error
}

// PlayerRepository provides persistence operations needed by background refreshes.
type PlayerRepository interface {
	GetSteamRefreshCandidates(context.Context, time.Time, int) ([]int, error)
	SaveSteamEnrichment(context.Context, int, *model.Player, time.Time, string) error
	NeedsProfileRefresh(context.Context, int, time.Time) (bool, error)
	SaveProfile(context.Context, *model.Player) (bool, error)
	Store(context.Context, *model.Player) error
	StoreAll(context.Context, []model.Player) error
	ExistsByID(ctx context.Context, id int) (bool, error)
	HasAnyRecord(ctx context.Context) (bool, error)
}

// TeamRepository provides persistence operations needed by background refreshes.
type TeamRepository interface {
	GetRefreshCandidates(context.Context, time.Time) ([]int, error)
	Store(context.Context, *model.Team) error
	Update(context.Context, *model.Team) error
	GetByID(ctx context.Context, id int) (*model.Team, error)
	ExistsByID(ctx context.Context, id int) (bool, error)
}

// LeagueDetailsRepository provides persistence operations needed by background refreshes.
type LeagueDetailsRepository interface {
	Store(context.Context, *model.LeagueDetails) error
	Update(context.Context, *model.LeagueDetails) error
	ExistsByID(ctx context.Context, id int) (bool, error)
	GetAllActiveForTiers(ctx context.Context, tiers []int) ([]model.LeagueDetails, error)
	GetByID(ctx context.Context, id int) (*model.LeagueDetails, error)
	UpdateLiveStatus(ctx context.Context, key int, newStatus bool) error
	UpdateTotalPrizePool(ctx context.Context, key int, prizePool int) error
	SetAllAsNotLive(ctx context.Context) error
}

// LeagueRepository provides persistence operations needed by background refreshes.
type LeagueRepository interface {
	Store(context.Context, *model.League) error
	StoreAll(context.Context, []model.League) error
	ExistsByID(ctx context.Context, id int) (bool, error)
	HasAnyRecord(ctx context.Context) (bool, error)
	GetAllActive(ctx context.Context) ([]model.LeagueDetails, error)
	GetMissingHistorical(ctx context.Context) ([]int, error)
}

// TeamRosterRepository provides persistence operations needed by background refreshes.
type TeamRosterRepository interface {
	GetByID(context.Context, int) (*model.TeamRoster, error)
	Store(context.Context, *model.TeamRoster) error
	Update(context.Context, *model.TeamRoster) error
	ExistsByTeamID(ctx context.Context, TeamID int) (bool, error)
}

// GameRepository provides persistence operations needed by background refreshes.
type GameRepository interface {
	StoreAll(ctx context.Context, games []model.Game) error
	GetAll(ctx context.Context) ([]model.Game, error)
	RemoveAll(ctx context.Context) error
}

// LeagueSeriesRepository provides persistence operations needed by background refreshes.
type LeagueSeriesRepository interface {
	ReplaceAllForLeague(ctx context.Context, leagueID int, series []model.SeriesInfo) error
}
