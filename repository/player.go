package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"time"
)

// PlayerRepository repository object
type PlayerRepository struct {
	Conn Database
}

// NewPlayerRepository creates new struct
func NewPlayerRepository(Conn Database) *PlayerRepository {
	return &PlayerRepository{Conn}
}

// Store store player model in db
func (pr *PlayerRepository) Store(ctx context.Context, player *model.Player) error {
	player.DBKey = strconv.Itoa(player.ID)
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := pr.Conn.Insert(ctx, "players", player)
	if e.ErrorCode(err) == e.ECONFLICT {
		return &e.Error{Op: "PlayerRepository.Store, record already exists", Err: err}
	} else if err != nil {
		return &e.Error{Op: "PlayerRepository.Store", Err: err}
	}

	return nil
}

// StoreAll store array of records in one batch
func (pr *PlayerRepository) StoreAll(ctx context.Context, players []model.Player) error {
	// set db keys for all elements
	for i, player := range players {
		players[i].DBKey = strconv.Itoa(player.ID)
	}

	// is 2 seconds enough?
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	err := pr.Conn.InsertMany(ctx, "players", players)
	if err != nil {
		return &e.Error{Op: "PlayerRepository.StoreAll", Err: err}
	}

	return nil
}

// ExistsByID check wether record exists in the DB
func (pr *PlayerRepository) ExistsByID(ctx context.Context, id int) (bool, error) {

	exists, err := existsInColByID(ctx, pr.Conn, "players", strconv.Itoa(id))
	if err != nil {
		return false, &e.Error{Op: "PlayerRepository.ExistsByID", Err: err}
	}

	return exists, nil
}

// HasAnyRecord return true if there are at least one record in the DB
func (pr *PlayerRepository) HasAnyRecord(ctx context.Context) (bool, error) {
	query := "RETURN LENGTH(FOR d IN players LIMIT 1 RETURN true) > 0"
	var exists bool

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := pr.Conn.Query(ctx, query, nil, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
		return false, &e.Error{Op: "PlayerRepository.HasAnyRecords", Err: err}
	}

	return exists, nil
}

// GetNames resolves only the requested profiles using their indexed document keys.
func (pr *PlayerRepository) GetNames(ctx context.Context, ids []int) (map[int]string, error) {
	names := make(map[int]string)
	if len(ids) == 0 {
		return names, nil
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	var rows []model.Player
	_, err := pr.Conn.Query(ctx, `RETURN (
  FOR id IN @ids
   LET player = DOCUMENT("players", TO_STRING(id))
   FILTER player != null
   RETURN {account_id: id, name: player.name}
 )`, map[string]any{"ids": ids}, &rows)
	if e.IsNotFound(err) {
		return names, nil
	}
	if err != nil {
		return nil, &e.Error{Op: "PlayerRepository.GetNames", Err: err}
	}
	for _, player := range rows {
		names[player.ID] = player.Name
	}
	return names, nil
}

// NeedsProfileRefresh keeps Steam fallbacks eligible for DPC upgrades, including after restart.
func (pr *PlayerRepository) NeedsProfileRefresh(ctx context.Context, id int, now time.Time) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	var due bool
	_, err := pr.Conn.Query(ctx, `LET p = DOCUMENT("players", @id)
 RETURN p == null || (p.profile_source == "steam" && (p.profile_checked_at == null || p.profile_checked_at <= @cutoff))`, map[string]any{"id": strconv.Itoa(id), "cutoff": now.Add(-24 * time.Hour).UnixMilli()}, &due)
	if e.IsNotFound(err) {
		return true, nil
	}
	return due, err
}

// SaveProfile atomically inserts or replaces only a Steam fallback. A concurrent
// Steam lookup can never overwrite a DPC profile. Return whether this is a new player.
func (pr *PlayerRepository) SaveProfile(ctx context.Context, player *model.Player) (bool, error) {
	if err := pr.Store(ctx, player); err == nil {
		return true, nil
	} else if e.ErrorCode(err) != e.ECONFLICT {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	err := pr.Conn.DoQueryBuilder(ctx, `FOR p IN players FILTER p._key == @key && p.profile_source == "steam"
 REPLACE p WITH MERGE(KEEP(p, "steam_name", "steam_location", "avatar_url", "steam_updated_at", "steam_next_refresh_at", "steam_status"), @player) IN players`, map[string]any{"key": player.DBKey, "player": player})
	return false, err
}

// GetSteamRefreshCandidates includes every valid player, regardless of DPC coverage or team activity.
func (pr *PlayerRepository) GetSteamRefreshCandidates(ctx context.Context, now time.Time, limit int) ([]int, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	var ids []int
	_, err := pr.Conn.Query(ctx, `RETURN (
 FOR p IN players
 FILTER p.account_id > 0 && p.account_id <= 4294967295
 FILTER p.steam_next_refresh_at == null || p.steam_next_refresh_at <= @now
 SORT p.steam_next_refresh_at ASC, p.account_id ASC
 LIMIT @limit RETURN p.account_id
 )`, map[string]any{"now": now.UnixMilli(), "limit": limit}, &ids)
	if e.IsNotFound(err) {
		return nil, nil
	}
	return ids, err
}

// SaveSteamEnrichment updates only Steam fields; failures preserve previously collected data.
func (pr *PlayerRepository) SaveSteamEnrichment(ctx context.Context, id int, profile *model.Player, now time.Time, status string) error {
	delay := 24 * time.Hour
	if status == "request_failed" {
		delay = time.Hour
	}
	patch := map[string]any{"steam_next_refresh_at": now.Add(delay).UnixMilli(), "steam_status": status}
	if profile != nil {
		patch["steam_name"] = profile.Name
		patch["steam_location"] = profile.SteamLocation
		patch["avatar_url"] = profile.AvatarURL
		patch["steam_updated_at"] = now.UnixMilli()
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	return pr.Conn.DoQueryBuilder(ctx, `FOR p IN players FILTER p._key == @key
 UPDATE p WITH MERGE(@patch, p.profile_source == "steam" && IS_STRING(@name) && @name != "" ? {name: @name} : {}) IN players`, map[string]any{"key": strconv.Itoa(id), "patch": patch, "name": patch["steam_name"]})
}
