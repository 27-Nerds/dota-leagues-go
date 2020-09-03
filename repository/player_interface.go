package repository

import "dota_league/model"

// PlayerRepositoryInterface inteface for user repository
type PlayerRepositoryInterface interface {
	Store(*model.Player) error
	StoreAll(*[]model.Player) error
	ExistsByID(id int) (bool, error)
	HasAnyRecord() (bool, error)
}
