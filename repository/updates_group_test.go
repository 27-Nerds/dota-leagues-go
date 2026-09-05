package repository

import (
	"dota_league/model"
	"testing"
	"time"
)

func TestUpdateGroupWindow(t *testing.T) {
	start := int64(1000)
	previous := model.Update{Entity: "roster", EntityID: 36, Action: "updated", GroupID: "first", GroupKind: "roster", GroupStartedAt: start, CreatedAt: start + 24*time.Hour.Milliseconds(), Changes: []model.UpdateChange{{Field: "members", Before: map[int]bool{1: true}, After: map[int]bool{}}}}
	next := model.Update{Entity: "roster", EntityID: 36, Action: "updated", CreatedAt: start + 48*time.Hour.Milliseconds(), Changes: []model.UpdateChange{{Field: "members", Before: map[string]any{}, After: map[string]any{"2": true}}}}
	if !canGroup(previous, next) {
		t.Fatal("boundary should group despite JSON map type differences")
	}
	next.CreatedAt++
	if canGroup(previous, next) {
		t.Fatal("window must not slide from latest observation")
	}
	next.CreatedAt--
	next.Changes[0].Before = map[string]any{"9": true}
	if canGroup(previous, next) {
		t.Fatal("discontinuous snapshots must remain separate")
	}
	next.Changes[0].Before = map[string]any{}
	next.EntityID++
	if canGroup(previous, next) {
		t.Fatal("different teams must remain separate")
	}
	next.EntityID--
	next.Action = "created"
	if canGroup(previous, next) {
		t.Fatal("initial imports must remain separate")
	}
	for _, tc := range []struct {
		entity, field, kind string
		window              time.Duration
	}{
		{"team", "wins", "record", time.Hour}, {"tournament", "total_prize_pool", "prize", time.Hour},
		{"team", "tag", "branding", 48 * time.Hour}, {"tournament", "start_timestamp", "schedule", 48 * time.Hour},
		{"player", "name", "", 0},
	} {
		kind, window := updateGroup(model.Update{Entity: tc.entity, Action: "updated", Changes: []model.UpdateChange{{Field: tc.field}}})
		if kind != tc.kind || window != tc.window {
			t.Fatalf("%s: %s %v", tc.field, kind, window)
		}
	}
}
func TestMergeUpdateChangesPreservesHistory(t *testing.T) {
	sources := []model.Update{
		{Changes: []model.UpdateChange{{Field: "name", Before: "A", After: "B"}, {Field: "wins", Before: 0, After: 1}}},
		{Changes: []model.UpdateChange{{Field: "name", Before: "B", After: "A"}, {Field: "wins", Before: 1, After: 2}}},
	}
	changes := mergeUpdateChanges(sources)
	if len(changes) != 1 || changes[0].Field != "wins" || changes[0].Before != 0 || changes[0].After != 2 {
		t.Fatalf("net changes: %+v", changes)
	}
	if sources[0].Changes[0].After != "B" || len(sources[0].Changes) != 2 {
		t.Fatal("raw history changed")
	}
}
