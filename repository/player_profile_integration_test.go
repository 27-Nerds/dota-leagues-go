package repository

import (
	"context"
	"dota_league/db"
	"dota_league/model"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSteamProfileUpgrade(t *testing.T) {
	endpoint := os.Getenv("ARANGO_TEST_URL")
	if endpoint == "" {
		t.Skip("set ARANGO_TEST_URL")
	}
	conn, err := db.Connect(t.Context(), endpoint, "", "", fmt.Sprintf("player_fallback_test_%d", time.Now().UnixNano()))
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
	repo := NewPlayerRepository(conn)
	now := time.Now()
	steam := &model.Player{ID: 42, Name: "Steam name", SteamLocation: "Bangladesh", AvatarURL: "https://avatars.steamstatic.com/test.jpg", ProfileSource: "steam", ProfileCheckedAt: now.UnixMilli()}
	if created, err := repo.SaveProfile(t.Context(), steam); err != nil || !created {
		t.Fatalf("initial save %v %v", created, err)
	}
	var cached model.Player
	if _, err := conn.Query(t.Context(), `RETURN DOCUMENT("players", "42")`, nil, &cached); err != nil {
		t.Fatal(err)
	}
	if cached.SteamLocation != "Bangladesh" || cached.CountryCode != "" {
		t.Fatalf("location not stored separately: %+v", cached)
	}
	for _, tc := range []struct {
		at   time.Time
		want bool
	}{{now, false}, {now.Add(25 * time.Hour), true}} {
		if due, err := repo.NeedsProfileRefresh(t.Context(), 42, tc.at); err != nil || due != tc.want {
			t.Fatalf("refresh %v %v", due, err)
		}
	}
	dpc := &model.Player{ID: 42, Name: "DPC name", IsPro: true, TeamID: 36}
	if created, err := repo.SaveProfile(t.Context(), dpc); err != nil || created {
		t.Fatalf("upgrade %v %v", created, err)
	}
	if _, err := repo.SaveProfile(t.Context(), steam); err != nil {
		t.Fatal(err)
	}
	var saved model.Player
	if _, err := conn.Query(t.Context(), `RETURN DOCUMENT("players", "42")`, nil, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Name != "DPC name" || saved.ProfileSource != "" || saved.AvatarURL != steam.AvatarURL || saved.SteamLocation != "Bangladesh" || !saved.IsPro || saved.TeamID != 36 {
		t.Fatalf("DPC replaced or stale Steam fields: %+v", saved)
	}
	ids, err := repo.GetSteamRefreshCandidates(t.Context(), now, 50)
	if err != nil || len(ids) != 1 || ids[0] != 42 {
		t.Fatalf("DPC player not eligible: %v %v", ids, err)
	}
	if err := repo.SaveSteamEnrichment(t.Context(), 42, &model.Player{Name: "Changed Steam name", SteamLocation: "Kyiv, Ukraine", AvatarURL: "https://avatars.steamstatic.com/new.jpg"}, now, "ok"); err != nil {
		t.Fatal(err)
	}
	ids, err = repo.GetSteamRefreshCandidates(t.Context(), now.Add(time.Hour), 50)
	if err != nil || len(ids) != 0 {
		t.Fatalf("fresh profile requeued: %v %v", ids, err)
	}
	if err := repo.SaveSteamEnrichment(t.Context(), 42, nil, now.Add(24*time.Hour), "request_failed"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Query(t.Context(), `RETURN DOCUMENT("players", "42")`, nil, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Name != "DPC name" || saved.TeamID != 36 || !saved.IsPro || saved.SteamName != "Changed Steam name" || saved.SteamLocation != "Kyiv, Ukraine" {
		t.Fatalf("enrichment corrupted profile: %+v", saved)
	}
	ids, err = repo.GetSteamRefreshCandidates(t.Context(), now.Add(25*time.Hour), 1)
	if err != nil || len(ids) != 1 {
		t.Fatalf("failed profile not retried: %v %v", ids, err)
	}
	if err := repo.SaveSteamEnrichment(t.Context(), 42, &model.Player{Name: "Steam name"}, now.Add(25*time.Hour), "ok"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Query(t.Context(), `RETURN DOCUMENT("players", "42")`, nil, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.SteamLocation != "" {
		t.Fatal("removed location retained")
	}
	if due, err := repo.NeedsProfileRefresh(t.Context(), 42, now.Add(48*time.Hour)); err != nil || due {
		t.Fatalf("DPC refresh %v %v", due, err)
	}
}
