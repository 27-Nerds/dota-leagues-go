package repository

import (
	"cmp"
	"context"
	"dota_league/db"
	"dota_league/model"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"
)

func TestUpdatesPersistenceAndPagination(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL to test against local ArangoDB")
	}
	conn, err := db.Connect(t.Context(), url, "", "", fmt.Sprintf("updates_test_%d", time.Now().UnixNano()))
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
	repo := NewUpdatesRepository(conn)
	rows, total, err := repo.GetAll(t.Context(), 0, 10)
	if err != nil || total != 0 || rows == nil {
		t.Fatalf("empty feed: %v %d %v", rows, total, err)
	}
	want := []model.Update{
		{Entity: "team", EntityID: 1, Name: "Team A", Action: "created", Changes: []model.UpdateChange{}},
		{Entity: "team", EntityID: 1, Name: "Team B", Action: "updated", Changes: []model.UpdateChange{{Field: "name", Before: "Team A", After: "Team B"}}},
		{Entity: "tournament", EntityID: 2, Action: "created", Changes: []model.UpdateChange{}},
	}
	for i := range want {
		if err := repo.Store(t.Context(), &want[i]); err != nil {
			t.Fatal(err)
		}
	}
	slices.SortFunc(want, func(a, b model.Update) int {
		if order := cmp.Compare(b.CreatedAt, a.CreatedAt); order != 0 {
			return order
		}
		return cmp.Compare(b.ID, a.ID)
	})
	// A new repository instance must see the stored feed, not process-local history.
	reader := NewUpdatesRepository(conn)
	for offset := range want {
		rows, total, err = reader.GetAll(t.Context(), offset, 1)
		if err != nil || total != 3 || len(rows) != 1 || rows[0].ID != want[offset].ID || rows[0].CreatedAt == 0 {
			t.Fatalf("page %d: %v %d %v", offset, rows, total, err)
		}
	}
	entry := &model.Update{Entity: "player", EntityID: 7, Team: &model.UpdateTeam{ID: 99, Name: "Recorded name"}}
	if err := repo.Store(t.Context(), entry); err != nil {
		t.Fatal(err)
	}
	checkAvailability := func(want bool) {
		t.Helper()
		rows, _, err := reader.GetAll(t.Context(), 0, 10)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			if row.ID == entry.ID {
				if row.Team == nil || row.Team.Available != want || row.Team.Name != "Recorded name" {
					t.Fatalf("team availability: %+v", row.Team)
				}
				return
			}
		}
		t.Fatal("player event missing")
	}
	checkAvailability(false)
	if err := NewTeamRepository(conn).Store(t.Context(), &model.Team{ID: 99, Name: "Current name"}); err != nil {
		t.Fatal(err)
	}
	// Group before pagination while retaining every original observation.
	first := &model.Update{Name: "Earlier roster name", Entity: "roster", EntityID: 36, Action: "updated", Changes: []model.UpdateChange{{Field: "members", Before: map[int]bool{1: true, 2: true}, After: map[int]bool{2: true}}}}
	second := &model.Update{Entity: "roster", EntityID: 36, Action: "updated", Changes: []model.UpdateChange{{Field: "members", Before: map[int]bool{2: true}, After: map[int]bool{2: true, 3: true}}}}
	for _, event := range []*model.Update{first, second} {
		if err := repo.Store(t.Context(), event); err != nil {
			t.Fatal(err)
		}
	}
	rows, total, err = reader.GetAll(t.Context(), 0, 1)
	if err != nil || total != 5 || len(rows) != 1 || len(rows[0].Sources) != 2 || rows[0].ID != second.ID {
		t.Fatalf("group page: %+v %d %v", rows, total, err)
	}
	if len(rows[0].Changes) != 1 || !equalUpdateValue(rows[0].Changes[0].Before, first.Changes[0].Before) || !equalUpdateValue(rows[0].Changes[0].After, second.Changes[0].After) {
		t.Fatalf("merged roster: %+v", rows[0].Changes)
	}
	var raw []model.Update
	_, err = conn.Query(t.Context(), `RETURN (FOR u IN updates FILTER u.entity == "roster" RETURN u)`, nil, &raw)
	if err != nil || len(raw) != 2 {
		t.Fatalf("raw history: %+v %v", raw, err)
	}
	for _, tc := range []struct {
		filter model.UpdateFilter
		count  int64
	}{
		{model.UpdateFilter{Entity: "roster", Search: "EARLIER", Days: 1}, 1},
		{model.UpdateFilter{Entity: "team", Search: "Earlier"}, 0},
		{model.UpdateFilter{Search: "recorded name"}, 1},
		{model.UpdateFilter{Search: "36"}, 1},
	} {
		found, count, err := reader.GetAll(t.Context(), 0, 1, tc.filter)
		if err != nil || count != tc.count || int64(len(found)) != tc.count {
			t.Fatalf("filter %+v: %v %d %v", tc.filter, found, count, err)
		}
		if tc.filter.Entity == "roster" && len(found[0].Sources) != 2 {
			t.Fatal("filter split a group")
		}
	}
	// Existing events gain a link once the profile is imported, without changing history.
	checkAvailability(true)
}
