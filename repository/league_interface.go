package repository

import (
	"dota_league/model"
)

// LeagueRepositoryInterface - interface for League repository
type LeagueRepositoryInterface interface {
	Store(*model.League) error
	StoreAll(*[]model.League) error
	ExistsByID(id int) (bool, error)
	HasAnyRecord() (bool, error)

	GetFromYearStart() (*[]model.League, error)
	GetByDateRange(startDate int64, endDate int64) (*[]model.League, error)
	GetAllActive() (*[]model.LeagueDetails, error)
}
