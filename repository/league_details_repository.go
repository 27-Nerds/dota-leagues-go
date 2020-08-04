package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"fmt"
	"strconv"
	"time"

	"github.com/arangodb/go-driver"
)

// LeagueDetailsRepository repository struct
type LeagueDetailsRepository struct {
	Conn           *db.Interface
	activityFilter string
}

// NewLeagueDetailsRepository creates new struct
func NewLeagueDetailsRepository(conn *db.Interface) LeagueDetailsRepositoryInterface {
	return &LeagueDetailsRepository{
		Conn:           conn,
		activityFilter: "FILTER (d.end_timestamp >= @today && d.status != 5) || d.is_live == true SORT d.tier DESC, d.is_live DESC, ABS(d.start_timestamp - @today), d.total_prize_pool DESC",
	}
}

// Store store leagueDetails model in db
func (ldr *LeagueDetailsRepository) Store(ld *model.LeagueDetails) error {
	ld.DbKey = strconv.Itoa(ld.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*ldr.Conn).Insert(ctx, "league_details", ld)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "LeagueDetailsRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.Store", Err: err}
	}

	return nil
}

// ExistsByID - check wether record exists in the DB
func (ldr *LeagueDetailsRepository) ExistsByID(id int) (bool, error) {

	exists, err := existsInColByID(ldr.Conn, "league_details", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "LeagueDetailsRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// GetAllActive returns all leagues wgere end_timestamp is greater than current date
func (ldr *LeagueDetailsRepository) GetAllActive() (*[]model.LeagueDetails, error) {

	query := fmt.Sprintf("FOR d IN league_details %s RETURN d", ldr.activityFilter)
	bindVars := map[string]interface{}{
		"today": time.Now().Unix(),
	}

	leagues, err := ldr.queryAll(query, bindVars)
	if err != nil {
		return nil, &e.Error{Op: "LeagueDetailsRepository.GetAllActive", Err: err}
	}

	return leagues, nil
}

// GetAllActiveForTiers returns all leagues wgere end_timestamp is greater than current date
func (ldr *LeagueDetailsRepository) GetAllActiveForTiers(tiers []int) (*[]model.LeagueDetails, error) {

	query := fmt.Sprintf("FOR d IN league_details FILTER d.tier in @tier %s RETURN d", ldr.activityFilter)
	bindVars := map[string]interface{}{
		"today": time.Now().Unix(),
		"tier":  tiers,
	}

	leagues, err := ldr.queryAll(query, bindVars)
	if err != nil {
		return nil, &e.Error{Op: "LeagueDetailsRepository.GetAllActiveForTiers", Err: err}
	}

	return leagues, nil
}

// UpdateLiveStatus you can set league as active or not
func (ldr *LeagueDetailsRepository) UpdateLiveStatus(key int, newStatus bool) error {
	patch := map[string]interface{}{
		"is_live": newStatus,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*ldr.Conn).Update(ctx, "league_details", strconv.Itoa(key), patch)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.UpdateLiveStatus", Err: err}
	}

	return nil
}

// UpdateTotalPrizePool set new prize pool for given league
func (ldr *LeagueDetailsRepository) UpdateTotalPrizePool(key int, prizePool int) error {
	patch := map[string]interface{}{
		"total_prize_pool": prizePool,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := (*ldr.Conn).Update(ctx, "league_details", strconv.Itoa(key), patch)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.UpdateTotalPrizePool", Err: err}
	}

	return nil
}

// SetAllAsNotLive set all leagues as inactive
func (ldr *LeagueDetailsRepository) SetAllAsNotLive() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	query := "FOR u IN league_details UPDATE u WITH { u.is_live: false } IN league_details"
	err := (*ldr.Conn).DoQuery(ctx, query)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.SetAllAsNotLive", Err: err}
	}

	return nil
}

// queryAll performs given query and returs array of serialized objects
func (ldr *LeagueDetailsRepository) queryAll(query string, bindVars map[string]interface{}) (*[]model.LeagueDetails, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cursor, err := (*ldr.Conn).QueryAll(ctx, query, bindVars)
	if err != nil {
		return nil, &e.Error{Op: "LeagueDetailsRepository.GetAllActive", Err: err}
	}

	defer cursor.Close()
	var leagues []model.LeagueDetails

	for {
		var doc model.LeagueDetails
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &e.Error{Op: "LeagueDetailsRepository.GetAllActive", Err: err}
		}
		leagues = append(leagues, doc)
	}

	return &leagues, nil
}
