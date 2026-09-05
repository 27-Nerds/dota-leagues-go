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
)

func TestLeagueArchiveViews(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL for database integration tests")
	}
	ctx := t.Context()
	conn, err := db.Connect(ctx, url, "", "", fmt.Sprintf("league_archive_%d", time.Now().UnixNano()))
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
	now := int(time.Now().Unix())
	details := NewLeagueDetailsRepository(conn)
	rows := []model.LeagueDetails{
		{ID: 1, Name: "Active", Tier: 2, Status: 3, EndTimestamp: now + 86400},
		{ID: 2, Name: "The International", Tier: 5, Status: 5, EndTimestamp: now - 86400},
		{ID: 3, Name: "Past amateur", Tier: 1, Status: 3, EndTimestamp: now - 86400},
		{ID: 4, Name: "Live overtime", Tier: 2, Status: 3, EndTimestamp: now - 86400, IsLive: true},
	}
	for i := range rows {
		if err := details.Store(ctx, &rows[i]); err != nil {
			t.Fatal(err)
		}
	}
	check := func(filter model.LeagueFilter, want []int) {
		t.Helper()
		got, total, err := details.GetAll(ctx, 0, 100, filter)
		if err != nil {
			t.Fatal(err)
		}
		ids := []int{}
		for _, l := range got {
			ids = append(ids, l.ID)
		}
		if !reflect.DeepEqual(ids, want) || total != int64(len(want)) {
			t.Fatalf("filter=%+v got=%v total=%d want=%v", filter, ids, total, want)
		}
	}
	check(model.LeagueFilter{}, []int{4, 1})
	check(model.LeagueFilter{Status: "completed"}, []int{2, 3})
	check(model.LeagueFilter{Status: "all"}, []int{4, 1, 2, 3})
	check(model.LeagueFilter{Status: "all", Sort: "name"}, []int{1, 4, 3, 2})
	check(model.LeagueFilter{Status: "all", Sort: "name", Order: "desc"}, []int{2, 3, 4, 1})
	check(model.LeagueFilter{Status: "all", Sort: "tier"}, []int{2, 1, 4, 3})
	check(model.LeagueFilter{Status: "all", Sort: "end_date", Order: "asc"}, []int{2, 3, 4, 1})
	tier := 5
	check(model.LeagueFilter{Status: "all", Tier: &tier, Search: " INTERNATIONAL "}, []int{2})
	check(model.LeagueFilter{Status: "completed", LiveOnly: true}, []int{})
	got, total, err := details.GetAll(ctx, 1, 1, model.LeagueFilter{Status: "all"})
	if err != nil || total != 4 || len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("archive pagination: %v %d %v", got, total, err)
	}
	leagues := NewLeagueRepository(conn)
	base := []model.League{{ID: 10, Name: "Missing TI", Tier: 5, Status: 5, EndTimestamp: int64(now - 86400)}, {ID: 11, Name: "Missing recent", Tier: 2, Status: 5, EndTimestamp: int64(now - 60)}}
	if err := leagues.StoreAll(ctx, base); err != nil {
		t.Fatal(err)
	}
	base[0].Name = "Updated TI"
	if err := leagues.StoreAll(ctx, base); err != nil {
		t.Fatal(err)
	}
	var name string
	if _, err := conn.Query(ctx, `RETURN DOCUMENT("leagues", "10").name`, nil, &name); err != nil || name != "Updated TI" {
		t.Fatalf("catalog upsert: %s %v", name, err)
	}
	ids, err := leagues.GetMissingHistorical(ctx)
	if err != nil || !reflect.DeepEqual(ids, []int{10, 11}) {
		t.Fatalf("TI backfill priority: %v %v", ids, err)
	}
	if err := details.Store(ctx, &model.LeagueDetails{ID: 10, Status: 5}); err != nil {
		t.Fatal(err)
	}
	ids, err = leagues.GetMissingHistorical(ctx)
	if err != nil || !reflect.DeepEqual(ids, []int{11}) {
		t.Fatalf("finished records must not be requeued: %v %v", ids, err)
	}
}
