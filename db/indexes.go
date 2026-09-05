package db

import (
	"context"
	"fmt"
	"log/slog"

	arango "github.com/arangodb/go-driver/v2/arangodb"
)

// Index definitions are shared by startup and the standalone maintenance command.
var collectionIndexes = map[string][][]string{
	"league_series":    {{"league_id", "start_time"}, {"start_time"}},
	"players":          {{"steam_next_refresh_at", "account_id"}},
	"updates":          {{"created_at", "id"}, {"entity", "entity_id", "group_kind", "created_at", "id"}},
	"update_groups":    {{"created_at", "id"}, {"entity", "created_at", "id"}},
	"source_snapshots": {{"kind", "entity_id", "first_seen"}},
}

// EnsureIndexes creates missing collections and indexes without changing records.
// It is safe to rerun and never drops or replaces an existing index.
func (a *ArangoDB) EnsureIndexes(ctx context.Context) error {
	for _, name := range []string{"league_series", "players", "updates", "update_groups", "source_snapshots"} {
		if err := a.ensureCollectionIndexes(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (a *ArangoDB) ensureCollectionIndexes(ctx context.Context, name string) error {
	col, err := a.collection(ctx, name)
	if err != nil {
		return err
	}
	background, disabled := true, false
	for _, fields := range collectionIndexes[name] {
		_, created, err := col.EnsurePersistentIndex(ctx, fields, &arango.CreatePersistentIndexOptions{
			InBackground: &background, Unique: &disabled, Sparse: &disabled,
		})
		if err != nil {
			return fmt.Errorf("ensure index %s(%v): %w", name, fields, err)
		}
		slog.Info("database index ready", "collection", name, "fields", fields, "created", created)
	}
	return nil
}
