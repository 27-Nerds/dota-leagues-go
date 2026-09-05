package db

import (
	"context"
	"time"
)

// EnsureUpdatesCollection prepares indexes and rebuilds derived feed groups
// before workers start. Rebuilding also repairs summaries after an older app
// version has written events. Original observations are never modified.
func (a *ArangoDB) EnsureUpdatesCollection(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	for _, name := range []string{"updates", "update_groups"} {
		if err := a.ensureCollectionIndexes(ctx, name); err != nil {
			return err
		}
	}
	return a.withTransaction(ctx, []string{"updates", "update_groups"}, true, func(txCtx context.Context) error {
		if err := a.ClearCollection(txCtx, "update_groups"); err != nil {
			return err
		}
		return a.DoQueryBuilder(txCtx, `FOR event IN updates
COLLECT group = event.group_id != null ? event.group_id : event.id INTO entries = event
LET ordered = (FOR item IN entries SORT item.created_at ASC, item.id ASC RETURN item)
LET latest = LAST(ordered)
INSERT {
 _key: group, latest: latest, entity: latest.entity, created_at: latest.created_at, id: latest.id,
 source_keys: ordered[*]._key,
 search_names: UNIQUE(FLATTEN(FOR item IN ordered RETURN [LOWER(TO_STRING(item.name)), LOWER(TO_STRING(item.team.name))])),
 entity_ids: UNIQUE(FOR item IN ordered RETURN TO_STRING(item.entity_id))
} INTO update_groups`, nil)
	})
}
