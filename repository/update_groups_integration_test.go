package repository

import (
	"context"
	"dota_league/db"
	"dota_league/model"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	arango "github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
)

// Retain the former implementation as an independent regression oracle.
const legacyUpdatesQuery = `FOR event IN updates
COLLECT group = event.group_id != null ? event.group_id : event.id INTO entries = event
LET ordered = (FOR item IN entries SORT item.created_at ASC, item.id ASC RETURN item)
LET latest = LAST(ordered)
FILTER @entity == "" || latest.entity == @entity
FILTER latest.created_at >= @since
FILTER @search == "" || LENGTH(
 FOR source IN ordered
 FILTER CONTAINS(LOWER(TO_STRING(source.name)), @search)
  || CONTAINS(LOWER(TO_STRING(source.team.name)), @search)
  || TO_STRING(source.entity_id) == @search
 RETURN 1
) > 0
SORT latest.created_at DESC, latest.id DESC LIMIT @offset, @limit
LET u = MERGE(latest, {sources: LENGTH(ordered) > 1 ? ordered : []})
LET team = u.team != null && u.team.id > 0 ? DOCUMENT("teams", TO_STRING(u.team.id)) : null
RETURN u.team == null ? u : MERGE(u, {team: MERGE(u.team, {available: team != null && team.team_id > 0})})`

func TestUpdateGroupsBackfillMatchesHistoricalFeed(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL to test against local ArangoDB")
	}
	conn, err := db.Connect(t.Context(), url, "", "", fmt.Sprintf("groups_test_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.DB.Remove(ctx); err != nil {
			t.Error(err)
		}
	})
	if err := conn.EnsureUpdatesCollection(t.Context()); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UnixMilli()
	events := []model.Update{
		{ID: "first", GroupID: "group", Entity: "team", EntityID: 1, Name: "Original name", CreatedAt: now - 2*86400000,
			Team: &model.UpdateTeam{ID: 1, Name: "Historical team"}, Changes: []model.UpdateChange{{Field: "name", Before: "A", After: "B"}}},
		{ID: "second", GroupID: "group", Entity: "team", EntityID: 1, Name: "Current name", CreatedAt: now,
			Changes: []model.UpdateChange{{Field: "name", Before: "B", After: "A"}}},
		{ID: "legacy", Entity: "player", EntityID: 2, Name: "Legacy singleton", CreatedAt: now - 100},
		{ID: "tie-a", GroupID: "tie", Entity: "roster", EntityID: 3, Name: "Earlier tied name", CreatedAt: now},
		{ID: "tie-z", GroupID: "tie", Entity: "roster", EntityID: 3, Name: "Latest tied name", CreatedAt: now},
	}
	for _, event := range events {
		if err := conn.Insert(t.Context(), "updates", event); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewUpdatesRepository(conn)
	for range 2 { // Restart/backfill must be repeatable, preserving raw observations.
		if err := conn.EnsureUpdatesCollection(t.Context()); err != nil {
			t.Fatal(err)
		}
		for _, filter := range []model.UpdateFilter{
			{}, {Days: 1}, {Search: "original", Days: 1}, {Search: "historical team"},
			{Search: "3"}, {Entity: "roster"}, {Search: "legacy"}, {Search: "absent"},
			{Entity: "player", Search: "original"},
		} {
			for _, offset := range []int{0, 1, 10} {
				got, total, err := repo.GetAll(t.Context(), offset, 1, filter)
				if err != nil {
					t.Fatal(err)
				}
				since := int64(0)
				if filter.Days > 0 {
					since = time.Now().Add(-24 * time.Hour).UnixMilli()
				}
				cursor, err := conn.QueryAll(t.Context(), legacyUpdatesQuery, map[string]any{
					"offset": offset, "limit": 1, "entity": filter.Entity, "search": filter.Search, "since": since,
				}, true)
				if err != nil {
					t.Fatal(err)
				}
				want := []model.Update{}
				for {
					var row model.Update
					_, err := cursor.ReadDocument(t.Context(), &row)
					if shared.IsNoMoreDocuments(err) {
						break
					}
					if err != nil {
						db.CloseCursor(cursor)
						t.Fatal(err)
					}
					if len(row.Sources) > 1 {
						row.Changes = mergeUpdateChanges(row.Sources)
					}
					want = append(want, row)
				}
				wantTotal := int64(cursor.Statistics().FullCountInt)
				db.CloseCursor(cursor)
				if total != wantTotal || !reflect.DeepEqual(got, want) {
					t.Fatalf("filter=%+v offset=%d: got %+v/%d want %+v/%d", filter, offset, got, total, want, wantTotal)
				}
			}
		}
	}
	// A summary write failure must roll back the original event insert too.
	col, err := conn.DB.GetCollection(t.Context(), "update_groups", nil)
	if err != nil {
		t.Fatal(err)
	}
	unique := true
	if _, _, err := col.EnsurePersistentIndex(t.Context(), []string{"entity"}, &arango.CreatePersistentIndexOptions{Unique: &unique}); err != nil {
		t.Fatal(err)
	}
	event := &model.Update{Entity: "team", EntityID: 99, Action: "created"}
	if err := repo.Store(t.Context(), event); err == nil {
		t.Fatal("expected summary unique-index conflict")
	}
	var count int
	if _, err := conn.Query(t.Context(), `RETURN LENGTH(FOR u IN updates FILTER u.id == @id RETURN 1)`, map[string]any{"id": event.ID}, &count); err != nil || count != 0 {
		t.Fatalf("event committed without its summary: count=%d err=%v", count, err)
	}
}
