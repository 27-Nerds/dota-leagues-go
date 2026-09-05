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

func TestTeamRefreshCandidates(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL for local database integration")
	}
	conn, err := db.Connect(t.Context(), url, "", "", fmt.Sprintf("refresh_test_%d", time.Now().UnixNano()))
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
	now := time.Now().Truncate(time.Second)
	repo := NewTeamRepository(conn)
	// 1: inactive due, 2: inactive fresh, 3: recent match due,
	// 4: recent match fresh, 5: empty roster recently changed, 6: upcoming match.
	for id, age := range map[int]time.Duration{1: 25 * time.Hour, 2: 3 * time.Hour, 3: 3 * time.Hour, 4: time.Hour, 5: 3 * time.Hour, 6: 3 * time.Hour} {
		if err := repo.Store(t.Context(), &model.Team{ID: id, UpdatedTimestamp: now.Add(-age).Unix()}); err != nil {
			t.Fatal(err)
		}
	}
	if err := NewLeagueSeriesRepository(conn).ReplaceAllForLeague(t.Context(), 1, []model.SeriesInfo{
		{SeriesID: 1, TeamID1: 3, TeamID2: 4, StartTime: now.Add(-time.Hour).Unix()},
		{SeriesID: 2, TeamID1: 6, StartTime: now.Add(24 * time.Hour).Unix()},
	}); err != nil {
		t.Fatal(err)
	}
	if err := conn.Insert(t.Context(), "games", map[string]any{"radiant_team_id": 0, "dire_team_id": 0}); err != nil {
		t.Fatal(err)
	}
	if err := conn.EnsureUpdatesCollection(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := conn.Insert(t.Context(), "updates", model.Update{Entity: "roster", EntityID: 5, Action: "updated", CreatedAt: now.Add(-time.Hour).UnixMilli()}); err != nil {
		t.Fatal(err)
	}
	ids, err := repo.GetRefreshCandidates(t.Context(), now)
	if err != nil || !reflect.DeepEqual(ids, []int{3, 5, 6, 1}) {
		t.Fatalf("due teams: %v, %v", ids, err)
	}
	// Exactly two hours is eligible, even if the team has no current players.
	if err := repo.Update(t.Context(), &model.Team{ID: 5, UpdatedTimestamp: now.Add(-2 * time.Hour).Unix()}); err != nil {
		t.Fatal(err)
	}
	ids, err = repo.GetRefreshCandidates(t.Context(), now)
	if err != nil || !reflect.DeepEqual(ids, []int{3, 6, 5, 1}) {
		t.Fatalf("boundary: %v, %v", ids, err)
	}
}
