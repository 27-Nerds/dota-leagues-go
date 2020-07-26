package repository

import "dota_league/model"

// GameRepositoryInterface - interface for game repository
type GameRepositoryInterface interface {
	ExistsByID(id int64) (bool, error)
	StoreAll(games *[]model.Game) error

	GetAll() (*[]model.Game, error)
	RemoveAll() error
}
