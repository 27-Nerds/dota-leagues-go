package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"strconv"
	"strings"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb/shared"
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

// GetByID returns one profile with its team and tournament names joined for display.
func (pr *PlayerRepository) GetByID(ctx context.Context, id int) (*model.Player, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	var player model.Player
	_, err := pr.Conn.Query(ctx, `LET p = DOCUMENT("players", @id)
FILTER p != null
LET team = p.team_id > 0 ? DOCUMENT("teams", TO_STRING(p.team_id)) : null
LET results = (
 FOR r IN (IS_ARRAY(p.results) ? p.results : [])
 LET details = DOCUMENT("league_details", TO_STRING(r.league_id))
 LET league = DOCUMENT("leagues", TO_STRING(r.league_id))
 RETURN MERGE(r, {
  league_name: details.name != null && details.name != "" ? details.name : league.name,
  league_available: details != null && details.league_id > 0
 })
)
LET history = (
 FOR h IN (IS_ARRAY(p.audit_entries) ? p.audit_entries : [])
 LET t = h.team_id > 0 ? DOCUMENT("teams", TO_STRING(h.team_id)) : null
 SORT h.start_timestamp DESC
 RETURN MERGE(h, {team_name: t.name != null && t.name != "" ? t.name : h.team_name, team_available: t != null && t.team_id > 0})
)
LET roster = FIRST(
 FOR t IN teams
 FILTER @account IN t.members[*].account_id
 FOR m IN t.members FILTER m.account_id == @account
 SORT m.time_joined DESC, t.team_id DESC LIMIT 1
 RETURN {id: t.team_id, name: t.name, tag: t.tag, joined_at: m.time_joined}
)
RETURN MERGE(p, {
 team_name: team.name != null && team.name != "" ? team.name : p.team_name,
 team_tag: team.tag != null && team.tag != "" ? team.tag : p.team_tag,
 team_available: team != null && team.team_id > 0,
 roster_team: roster != null && roster.id > 0 ? roster : null,
 audit_entries: history,
 results
})`, map[string]any{"id": strconv.Itoa(id), "account": id}, &player)
	if err != nil {
		return nil, &e.Error{Op: "PlayerRepository.GetByID", Err: err}
	}
	return &player, nil
}

// GetAll lists the player directory. Professionals with the highest recorded earnings
// rank first by default; name and earnings sorts are explicit. ID breaks ties for stable pagination.
func (pr *PlayerRepository) GetAll(ctx context.Context, offset, limit int, filter model.PlayerFilter) ([]model.Player, int64, error) {
	sort := "listed DESC, (p.team_id > 0) DESC, LOWER(p.name) ASC, p.account_id ASC"
	if field, ok := map[string]string{"name": "LOWER(p.name)"}[filter.Sort]; ok {
		sort = field + " " + sortOrder(filter.Sort, filter.Order) + ", p.account_id ASC"
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	// "Professional" means the account appears in Valve's pro player feed; Steam-only
	// fallbacks are the rest. Valve no longer sends an is_pro flag.
	cursor, err := pr.Conn.QueryAll(ctx, `FOR p IN players
 FILTER p.account_id > 0
 LET listed = p.profile_source != "steam"
 FILTER @search == "" || CONTAINS(LOWER(p.name), @search) || CONTAINS(LOWER(p.real_name), @search) || CONTAINS(LOWER(p.steam_name), @search) || TO_STRING(p.account_id) == @search
 FILTER @country == "" || UPPER(p.country_code) == @country
 FILTER @pro == null || listed == @pro
 FILTER @hasTeam == null || (p.team_id > 0) == @hasTeam
 SORT `+sort+`
 LIMIT @offset, @limit
 LET team = p.team_id > 0 ? DOCUMENT("teams", TO_STRING(p.team_id)) : null
 RETURN MERGE(p, {team_name: team.name != null && team.name != "" ? team.name : p.team_name, team_tag: team.tag != null && team.tag != "" ? team.tag : p.team_tag})`,
		map[string]any{"offset": offset, "limit": limit, "search": strings.ToLower(strings.TrimSpace(filter.Search)), "country": strings.ToUpper(strings.TrimSpace(filter.Country)), "pro": filter.Pro, "hasTeam": filter.HasTeam}, true)
	if shared.IsNotFound(err) || e.IsNotFound(err) {
		return []model.Player{}, 0, nil
	}
	if err != nil {
		return nil, 0, &e.Error{Op: "PlayerRepository.GetAll", Err: err}
	}
	defer db.CloseCursor(cursor)
	players := []model.Player{}
	for {
		var player model.Player
		if _, err := cursor.ReadDocument(ctx, &player); shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, 0, &e.Error{Op: "PlayerRepository.GetAll", Err: err}
		}
		player.DBKey, player.SteamNextRefreshAt, player.ProfileCheckedAt = "", 0, 0
		players = append(players, player)
	}
	return players, int64(cursor.Statistics().FullCountInt), nil
}

// GetProfiles returns the stored documents for the requested IDs; missing players are omitted.
func (pr *PlayerRepository) GetProfiles(ctx context.Context, ids []int) (map[int]model.Player, error) {
	players := make(map[int]model.Player, len(ids))
	if len(ids) == 0 {
		return players, nil
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	cursor, err := pr.Conn.QueryAll(ctx, `FOR id IN @ids
 LET p = DOCUMENT("players", TO_STRING(id))
 FILTER p != null RETURN p`, map[string]any{"ids": ids}, false)
	if shared.IsNotFound(err) || e.IsNotFound(err) {
		return players, nil
	}
	if err != nil {
		return nil, &e.Error{Op: "PlayerRepository.GetProfiles", Err: err}
	}
	defer db.CloseCursor(cursor)
	for {
		var player model.Player
		if _, err := cursor.ReadDocument(ctx, &player); shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, &e.Error{Op: "PlayerRepository.GetProfiles", Err: err}
		}
		players[player.ID] = player
	}
	return players, nil
}

// GetSitemapPlayers lists IDs of players from Valve's pro feed in a stable order for sitemap pages.
func (pr *PlayerRepository) GetSitemapPlayers(ctx context.Context, offset, limit int) ([]int, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	cursor, err := pr.Conn.QueryAll(ctx, `FOR p IN players
 FILTER p.profile_source != "steam" && p.account_id > 0
 SORT p.account_id ASC LIMIT @offset, @limit RETURN p.account_id`, map[string]any{"offset": offset, "limit": limit}, true)
	if shared.IsNotFound(err) || e.IsNotFound(err) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, &e.Error{Op: "PlayerRepository.GetSitemapPlayers", Err: err}
	}
	defer db.CloseCursor(cursor)
	ids := []int{}
	for {
		var id int
		if _, err := cursor.ReadDocument(ctx, &id); shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return nil, 0, &e.Error{Op: "PlayerRepository.GetSitemapPlayers", Err: err}
		}
		ids = append(ids, id)
	}
	return ids, int64(cursor.Statistics().FullCountInt), nil
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

// SaveProfile atomically inserts a player or replaces the stored identity. A DPC
// profile always refreshes the record; a Steam fallback replaces only another
// Steam fallback, so a concurrent Steam lookup can never overwrite a DPC profile.
// Steam enrichment fields survive either replacement. Return whether this is a new player.
func (pr *PlayerRepository) SaveProfile(ctx context.Context, player *model.Player) (bool, error) {
	if err := pr.Store(ctx, player); err == nil {
		return true, nil
	} else if e.ErrorCode(err) != e.ECONFLICT {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	err := pr.Conn.DoQueryBuilder(ctx, `FOR p IN players FILTER p._key == @key && (p.profile_source == "steam" || @source != "steam")
 REPLACE p WITH MERGE(KEEP(p, "steam_name", "steam_location", "avatar_url", "steam_updated_at", "steam_next_refresh_at", "steam_status", "steam_privacy", "steam_checked_at", "steam_public_at"), @player) IN players`, map[string]any{"key": player.DBKey, "player": player, "source": player.ProfileSource})
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
	patch := map[string]any{"steam_next_refresh_at": now.Add(delay).UnixMilli(), "steam_status": status, "steam_checked_at": now.UnixMilli()}
	if profile != nil {
		patch["steam_name"] = profile.Name
		patch["avatar_url"] = profile.AvatarURL
		patch["steam_updated_at"] = now.UnixMilli()
		patch["steam_privacy"] = profile.SteamPrivacy
		// A private profile hides its location; keep the last public value instead of blanking it.
		if profile.SteamLocation != "" || profile.SteamPrivacy == "" || profile.SteamPrivacy == "public" {
			patch["steam_location"] = profile.SteamLocation
		}
		if profile.SteamPrivacy == "" || profile.SteamPrivacy == "public" {
			patch["steam_public_at"] = now.UnixMilli()
		}
	}
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	return pr.Conn.DoQueryBuilder(ctx, `FOR p IN players FILTER p._key == @key
 UPDATE p WITH MERGE(@patch, p.profile_source == "steam" && IS_STRING(@name) && @name != "" ? {name: @name} : {}) IN players`, map[string]any{"key": strconv.Itoa(id), "patch": patch, "name": patch["steam_name"]})
}
