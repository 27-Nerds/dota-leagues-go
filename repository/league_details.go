package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"fmt"
	"strconv"
	"strings"
	"time"

	driver "github.com/arangodb/go-driver/v2/arangodb/shared"
)

// LeagueDetailsRepository repository struct
type LeagueDetailsRepository struct {
	Conn           Database
	activityFilter string
}

// NewLeagueDetailsRepository creates new struct
func NewLeagueDetailsRepository(conn Database) *LeagueDetailsRepository {
	return &LeagueDetailsRepository{
		Conn:           conn,
		activityFilter: "FILTER (d.end_timestamp >= @today && d.status != 5) || d.is_live == true SORT d.tier DESC, d.is_live DESC, ABS(d.start_timestamp - @today), d.total_prize_pool DESC",
	}
}

// Store store leagueDetails model in db
func (ldr *LeagueDetailsRepository) Store(ctx context.Context, ld *model.LeagueDetails) error {
	ld.DBKey = strconv.Itoa(ld.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := ldr.Conn.Insert(ctx, "league_details", ld)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "LeagueDetailsRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.Store", Err: err}
	}

	return nil
}

// Update updates an existing league details record (partial merge)
func (ldr *LeagueDetailsRepository) Update(ctx context.Context, ld *model.LeagueDetails) error {
	ld.DBKey = strconv.Itoa(ld.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := ldr.Conn.Update(ctx, "league_details", ld.DBKey, ld)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.Update", Err: err}
	}

	return nil
}

// ExistsByID - check wether record exists in the DB
func (ldr *LeagueDetailsRepository) ExistsByID(ctx context.Context, id int) (bool, error) {

	exists, err := existsInColByID(ctx, ldr.Conn, "league_details", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "LeagueDetailsRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// GetAll returns a paginated lifecycle view with search and tournament filters.
func (ldr *LeagueDetailsRepository) GetAll(ctx context.Context, offset int, limit int, filter model.LeagueFilter) ([]model.LeagueDetails, int64, error) {
	sort := "active DESC, d.tier DESC, d.is_live DESC, (active ? ABS(d.start_timestamp - @today) : -d.end_timestamp), d.total_prize_pool DESC, d.league_id ASC"
	if field, ok := map[string]string{"name": "LOWER(d.name)", "start_date": "d.start_timestamp", "end_date": "d.end_timestamp", "prize_pool": "d.total_prize_pool", "tier": "d.tier"}[filter.Sort]; ok {
		sort = field + " " + sortOrder(filter.Sort, filter.Order) + ", d.league_id ASC"
	}

	query := `FOR d IN league_details
 LET active = (d.end_timestamp >= @today && d.status != 5) || d.is_live == true
 FILTER @status == "all" || (@status == "completed" ? !active : active)
 FILTER @search == "" || CONTAINS(LOWER(d.name), @search)
 FILTER @tier == null || d.tier == @tier
 FILTER @region == null || d.region == @region
 FILTER !@liveOnly || d.is_live == true
 SORT ` + sort + `
 LIMIT @offset, @limit RETURN d`
	bindVars := map[string]any{
		"today": time.Now().Unix(), "status": filter.Status, "search": strings.ToLower(strings.TrimSpace(filter.Search)),
		"tier": filter.Tier, "region": filter.Region, "liveOnly": filter.LiveOnly, "offset": offset, "limit": limit,
	}

	leagues, totalCount, err := ldr.queryAll(ctx, query, bindVars, true)
	if err != nil {
		return nil, 0, &e.Error{Op: "LeagueDetailsRepository.GetAllActive", Err: err}
	}

	return leagues, totalCount, nil
}

// GetAllActiveForTiers returns all leagues wgere end_timestamp is greater than current date
func (ldr *LeagueDetailsRepository) GetAllActiveForTiers(ctx context.Context, tiers []int) ([]model.LeagueDetails, error) {

	query := fmt.Sprintf("FOR d IN league_details FILTER d.tier in @tier %s RETURN d", ldr.activityFilter)
	bindVars := map[string]any{
		"today": time.Now().Unix(),
		"tier":  tiers,
	}

	leagues, _, err := ldr.queryAll(ctx, query, bindVars, false)
	if err != nil {
		return nil, &e.Error{Op: "LeagueDetailsRepository.GetAllActiveForTiers", Err: err}
	}

	return leagues, nil
}

// UpdateLiveStatus you can set league as active or not
func (ldr *LeagueDetailsRepository) UpdateLiveStatus(ctx context.Context, key int, newStatus bool) error {
	patch := map[string]any{
		"is_live": newStatus,
	}

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := ldr.Conn.Update(ctx, "league_details", strconv.Itoa(key), patch)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.UpdateLiveStatus", Err: err}
	}

	return nil
}

// UpdateTotalPrizePool set new prize pool for given league
func (ldr *LeagueDetailsRepository) UpdateTotalPrizePool(ctx context.Context, key int, prizePool int) error {
	patch := map[string]any{
		"total_prize_pool": prizePool,
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := ldr.Conn.Update(ctx, "league_details", strconv.Itoa(key), patch)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.UpdateTotalPrizePool", Err: err}
	}

	return nil
}

// SetAllAsNotLive set all leagues as inactive
func (ldr *LeagueDetailsRepository) SetAllAsNotLive(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	query := "FOR u IN league_details UPDATE u WITH { is_live: false } IN league_details"
	err := ldr.Conn.DoQuery(ctx, query)
	if err != nil {
		return &e.Error{Op: "LeagueDetailsRepository.SetAllAsNotLive", Err: err}
	}

	return nil
}

// GetByID get league
func (ldr *LeagueDetailsRepository) GetByID(ctx context.Context, id int) (*model.LeagueDetails, error) {
	query := "FOR d IN league_details FILTER d._key == @id RETURN d"
	bindVars := map[string]any{
		"id": strconv.Itoa(id),
	}

	var league model.LeagueDetails

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := ldr.Conn.Query(ctx, query, bindVars, &league)

	if err != nil {
		return nil, &e.Error{Op: "LeagueDetailsRepository.Get", Err: err}
	}

	return &league, nil
}

// queryAll performs given query and returs array of serialized objects
// second return parameter is total count of results, if withTotalCount is set to false, it will be 0
func (ldr *LeagueDetailsRepository) queryAll(ctx context.Context, query string, bindVars map[string]any, withTotalCount bool) ([]model.LeagueDetails, int64, error) {
	var totalCount int64

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	cursor, err := ldr.Conn.QueryAll(ctx, query, bindVars, withTotalCount)
	if err != nil {
		return nil, totalCount, &e.Error{Op: "LeagueDetailsRepository.GetAllActive", Err: err}
	}

	defer db.CloseCursor(cursor)
	leagues := []model.LeagueDetails{}

	for {
		var doc model.LeagueDetails
		_, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, totalCount, &e.Error{Op: "LeagueDetailsRepository.GetAllActive", Err: err}
		}
		leagues = append(leagues, doc)
	}
	if withTotalCount {
		totalCount = int64(cursor.Statistics().FullCountInt)
	}

	return leagues, totalCount, nil
}
