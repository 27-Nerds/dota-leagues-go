package repository

import (
	"bytes"
	"context"
	"crypto/rand"
	"dota_league/db"
	e "dota_league/error"
	"dota_league/model"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/arangodb/go-driver/v2/arangodb/shared"
)

type UpdatesRepository struct {
	Conn   Database
	mu     sync.Mutex
	lastAt int64
}

func NewUpdatesRepository(conn Database) *UpdatesRepository {
	return &UpdatesRepository{Conn: conn}
}

func (r *UpdatesRepository) Store(ctx context.Context, update *model.Update) error {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	r.mu.Lock()
	defer r.mu.Unlock()
	update.ID = rand.Text()
	update.CreatedAt = max(time.Now().UnixMilli(), r.lastAt+1)
	r.lastAt = update.CreatedAt
	update.GroupID = update.ID
	update.GroupStartedAt = update.CreatedAt
	update.GroupKind, _ = updateGroup(*update)
	if update.GroupKind != "" {
		var previous model.Update
		_, err := r.Conn.Query(ctx, `FOR u IN updates
   FILTER u.entity == @entity && u.entity_id == @entityID && u.group_kind == @kind
   SORT u.created_at DESC, u.id DESC LIMIT 1 RETURN u`,
			map[string]any{"entity": update.Entity, "entityID": update.EntityID, "kind": update.GroupKind}, &previous)
		if err != nil && !e.IsNotFound(err) && !shared.IsNotFound(err) {
			return err
		}
		if err == nil && canGroup(previous, *update) {
			update.GroupID = previous.GroupID
			update.GroupStartedAt = previous.GroupStartedAt
		}
	}
	// Both writes commit together, so readers never see a partial group update.
	return r.Conn.WithTransaction(ctx, "updates", func(txCtx context.Context) error {
		return r.Conn.DoQueryBuilder(txCtx, `INSERT @event INTO updates
LET saved = NEW
LET names = [LOWER(TO_STRING(saved.name)), LOWER(TO_STRING(saved.team.name))]
UPSERT {_key: @group}
INSERT {_key: @group, latest: saved, entity: saved.entity, created_at: saved.created_at, id: saved.id,
 source_keys: [saved._key], search_names: UNIQUE(names), entity_ids: [TO_STRING(saved.entity_id)]}
UPDATE MERGE({
 source_keys: APPEND(OLD.source_keys, saved._key),
 search_names: UNION_DISTINCT(OLD.search_names, names),
 entity_ids: UNION_DISTINCT(OLD.entity_ids, [TO_STRING(saved.entity_id)])},
 [saved.created_at, saved.id] > [OLD.created_at, OLD.id]
  ? {latest: saved, entity: saved.entity, created_at: saved.created_at, id: saved.id} : {})
IN update_groups`, map[string]any{"event": update, "group": update.GroupID})
	}, "update_groups")
}

func (r *UpdatesRepository) GetAll(ctx context.Context, offset, limit int, filters ...model.UpdateFilter) ([]model.Update, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	filter := model.UpdateFilter{}
	if len(filters) > 0 {
		filter = filters[0]
	}
	since := int64(0)
	if filter.Days > 0 && filter.Days <= 30 {
		since = time.Now().Add(-time.Duration(filter.Days) * 24 * time.Hour).UnixMilli()
	}
	rows := []model.Update{}
	cursor, err := r.Conn.QueryAll(ctx,
		`FOR g IN update_groups
FILTER g.created_at >= @since
FILTER @entity == "" || g.entity == @entity
FILTER @search == "" || @search IN g.entity_ids || LENGTH(
 FOR name IN g.search_names FILTER CONTAINS(name, @search) LIMIT 1 RETURN 1
) > 0
SORT g.created_at DESC, g.id DESC LIMIT @offset, @limit
LET ordered = LENGTH(g.source_keys) > 1 ? (
 FOR key IN g.source_keys
 LET source = DOCUMENT("updates", key)
 SORT source.created_at ASC, source.id ASC RETURN source
) : []
LET u = MERGE(g.latest, {sources: ordered})
LET team = u.team != null && u.team.id > 0 ? DOCUMENT("teams", TO_STRING(u.team.id)) : null
RETURN u.team == null ? u : MERGE(u, {team: MERGE(u.team, {available: team != null && team.team_id > 0})})`,
		map[string]any{"offset": offset, "limit": limit, "entity": filter.Entity, "search": strings.ToLower(strings.TrimSpace(filter.Search)), "since": since}, true)
	if shared.IsNotFound(err) || e.IsNotFound(err) {
		return rows, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	defer db.CloseCursor(cursor)
	for {
		var row model.Update
		_, err := cursor.ReadDocument(ctx, &row)
		if shared.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		if len(row.Sources) > 1 {
			row.Changes = mergeUpdateChanges(row.Sources)
		}
		rows = append(rows, row)
	}
	return rows, int64(cursor.Statistics().FullCountInt), nil
}

// Windows are anchored to the first observation, never extended by later updates.
func updateGroup(u model.Update) (string, time.Duration) {
	if u.Action != "updated" || len(u.Changes) == 0 {
		return "", 0
	}
	if u.Entity == "roster" {
		return "roster", 48 * time.Hour
	}
	allowed := func(fields ...string) bool {
		for _, c := range u.Changes {
			found := false
			for _, f := range fields {
				if c.Field == f {
					found = true
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	if u.Entity == "team" {
		if allowed("wins", "losses") {
			return "record", time.Hour
		}
		if allowed("name", "tag", "url", "url_logo") {
			return "branding", 48 * time.Hour
		}
	}
	if u.Entity == "tournament" {
		if allowed("start_timestamp", "end_timestamp", "status") {
			return "schedule", 48 * time.Hour
		}
		if allowed("total_prize_pool") {
			return "prize", time.Hour
		}
	}
	if u.Entity == "player" {
		if allowed("total_earnings") {
			return "earnings", 24 * time.Hour
		}
		if allowed("steam_name", "steam_location", "steam_profile") {
			return "steam", 24 * time.Hour
		}
	}
	return "", 0
}
func equalUpdateValue(a, b any) bool {
	left, err := json.Marshal(a)
	if err != nil {
		return false
	}
	right, err := json.Marshal(b)
	return err == nil && bytes.Equal(left, right)
}
func canGroup(previous, next model.Update) bool {
	kind, window := updateGroup(next)
	if kind == "" || previous.GroupKind != kind || previous.GroupID == "" ||
		previous.Entity != next.Entity || previous.EntityID != next.EntityID ||
		next.CreatedAt < previous.CreatedAt || next.CreatedAt-previous.GroupStartedAt > window.Milliseconds() {
		return false
	}
	for _, a := range previous.Changes {
		for _, b := range next.Changes {
			if a.Field == b.Field && !equalUpdateValue(a.After, b.Before) {
				return false
			}
		}
	}
	return true
}
func mergeUpdateChanges(sources []model.Update) []model.UpdateChange {
	changes := []model.UpdateChange{}
	positions := map[string]int{}
	for _, source := range sources {
		for _, c := range source.Changes {
			if i, ok := positions[c.Field]; ok {
				changes[i].After = c.After
			} else {
				positions[c.Field] = len(changes)
				changes = append(changes, c)
			}
		}
	}
	result := []model.UpdateChange{}
	for _, c := range changes {
		if !equalUpdateValue(c.Before, c.After) {
			result = append(result, c)
		}
	}
	return result
}
