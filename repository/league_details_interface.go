package repository

import (
	"dota_league/model"
)

// LeagueDetailsRepositoryInterface - interface for LeagueDetails repository
type LeagueDetailsRepositoryInterface interface {
	Store(*model.LeagueDetails) error
	ExistsByID(id int) (bool, error)
	GetAllActive() (*[]model.LeagueDetails, error)
	GetAllActiveForTiers(tiers []int) (*[]model.LeagueDetails, error)
	UpdateLiveStatus(key int, newStatus bool) error
	UpdateTotalPrizePool(key int, prizePool int) error
	SetAllAsNotLive() error
}
