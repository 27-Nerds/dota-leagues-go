package repository

import "dota_league/model"

// LiveGameDetailsInterface inteface for live game repository
type LiveGameDetailsInterface interface {
	Store(*model.LiveGameDetails) error
	ExistsByID(id int64) (bool, error)
	Update(lgm *model.LiveGameDetails) error
}
