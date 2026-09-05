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

func TestMatchSitemapDeduplicatesAndPaginates(t *testing.T) {
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
	if err := repo.ReplaceAllForLeague(ctx, 1, []model.SeriesInfo{{SeriesID: 1, MatchIDs: []string{"123", "123", "456", "0", "bad", "01"}}, {SeriesID: 2, MatchIDs: []string{"123"}}}); err != nil {
		t.Fatal(err)
	}
	for offset, want := range []string{"123", "456"} {
		rows, total, err := repo.GetSitemapMatches(ctx, offset, 1)
		if err != nil || total != 2 || !reflect.DeepEqual(rows, []model.MatchReference{{LeagueID: 1, MatchID: want}}) {
			t.Fatalf("rows=%v total=%d err=%v", rows, total, err)
		}
	}
	rows, _, err := repo.GetSitemapMatches(ctx, 2, 1)
	if err != nil || len(rows) != 0 {
		t.Fatalf("empty page: %v %v", rows, err)
	}
}
