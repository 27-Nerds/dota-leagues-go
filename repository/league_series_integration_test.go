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

// Uses a disposable database; regular unit-test runs need no ArangoDB instance.
func TestLeagueSeriesReplacementIsAtomic(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL to run against a local ArangoDB instance")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := db.Connect(ctx, url, "", "", fmt.Sprintf("quality_review_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.DB.Remove(cleanupCtx); err != nil {
			t.Errorf("remove temporary database: %v", err)
		}
	})
	repo := NewLeagueSeriesRepository(conn)
	assertSeries := func(leagueID int, want []int) {
		t.Helper()
		var got []int
		_, err := conn.Query(ctx,
			"RETURN (FOR s IN league_series FILTER s.league_id == @id SORT s.series_id RETURN s.series_id)",
			map[string]any{"id": leagueID}, &got)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("league %d: got %v / %v, want %v", leagueID, got, err, want)
		}
	}
	if err := repo.ReplaceAllForLeague(t.Context(), 1, []model.SeriesInfo{{SeriesID: 1, StartTime: 2, TeamID1: 7}, {SeriesID: 2, StartTime: 1}}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceAllForLeague(t.Context(), 2, []model.SeriesInfo{{SeriesID: 9}}); err != nil {
		t.Fatal(err)
	}
	// Exercise v2 single-document updates, joins, and full-count pagination.
	teams := NewTeamRepository(conn)
	team := &model.Team{ID: 7, Name: "Before"}
	if err := teams.Store(t.Context(), team); err != nil {
		t.Fatal(err)
	}
	team.Name = "After"
	if err := teams.Update(t.Context(), team); err != nil {
		t.Fatal(err)
	}
	page, total, err := repo.GetAllByLeague(t.Context(), 1, 0, 1)
	if err != nil || total != 2 || len(page) != 1 || page[0].TeamName1 != "After" {
		t.Fatalf("paginated series: page=%v total=%d err=%v", page, total, err)
	}
	// A duplicate snapshot must still take the update path with v2 error helpers.
	standings := NewDPCStandingsRepository(conn)
	for _, timestamp := range []int64{1, 2} {
		if err := standings.Store(t.Context(), &model.DPCStandings{UpdatedTimestamp: timestamp}); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := standings.Get(t.Context())
	if err != nil || snapshot == nil || snapshot.UpdatedTimestamp != 2 {
		t.Fatalf("standings upsert: snapshot=%v err=%v", snapshot, err)
	}

	// Duplicate keys force an insert failure after deletion inside the transaction.
	if err := repo.ReplaceAllForLeague(t.Context(), 1, []model.SeriesInfo{{SeriesID: 3}, {SeriesID: 3}}); err == nil {
		t.Fatal("expected duplicate-key failure")
	}
	assertSeries(1, []int{1, 2})
	assertSeries(2, []int{9})
	if err := repo.ReplaceAllForLeague(t.Context(), 1, []model.SeriesInfo{{SeriesID: 4}}); err != nil {
		t.Fatal(err)
	}
	assertSeries(1, []int{4})
	if err := repo.ReplaceAllForLeague(t.Context(), 1, nil); err != nil {
		t.Fatal(err)
	}
	assertSeries(1, []int{})
	assertSeries(2, []int{9})
}
