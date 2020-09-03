package repository

import "dota_league/model"

// TeamRepositoryInterface inteface for user repository
type TeamRepositoryInterface interface {
	Store(*model.Team) error
	ExistsByID(id int) (bool, error)
}
