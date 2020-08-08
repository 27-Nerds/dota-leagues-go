package repository

import (
	"dota_league/model"
)

// LeagueDetailsRepositoryInterface - interface for LeagueDetails repository
type LeagueDetailsRepositoryInterface interface {
	Store(*model.LeagueDetails) error
	ExistsByID(id int) (bool, error)
	GetAllActive(offset int, limit int) (*[]model.LeagueDetails, int64, error)
	GetAllActiveForTiers(tiers []int) (*[]model.LeagueDetails, error)
	UpdateLiveStatus(key int, newStatus bool) error
	UpdateTotalPrizePool(key int, prizePool int) error
	SetAllAsNotLive() error
}
