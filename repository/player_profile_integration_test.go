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
	// A later DPC sync must refresh the professional identity while keeping Steam data.
	if err := NewTeamRepository(conn).Store(t.Context(), &model.Team{ID: 37, Name: "New Team", Tag: "NT"}); err != nil {
		t.Fatal(err)
	}
	moved := &model.Player{ID: 42, Name: "DPC name", IsPro: true, TeamID: 37, Results: []model.PlayerResult{{LeagueID: 1, Placement: 2}}}
	if created, err := repo.SaveProfile(t.Context(), moved); err != nil || created {
		t.Fatalf("DPC refresh %v %v", created, err)
	}
	profile, err := repo.GetByID(t.Context(), 42)
	if err != nil || profile.TeamID != 37 || profile.TeamName != "New Team" || !profile.TeamAvailable || profile.SteamName != "Steam name" || len(profile.Results) != 1 || profile.Results[0].LeagueAvailable || profile.RosterTeam != nil {
		t.Fatalf("profile read: %+v %v", profile, err)
	}
	// The Valve roster can list a player the DPC feed shows without a team; the newest roster entry wins.
	if _, err := conn.Query(t.Context(), `LET rows = [
  {_key: "38", team_id: 38, name: "Roster Team", tag: "RT", updated_timestamp: 300, members: [{account_id: 42, time_joined: 200}]},
  {_key: "39", team_id: 39, name: "Old Team", members: [{account_id: 42, time_joined: 100}]}]
 FOR r IN rows INSERT r INTO teams RETURN 1`, nil, new(int)); err != nil {
		t.Fatal(err)
	}
	profile, err = repo.GetByID(t.Context(), 42)
	if err != nil || profile.RosterTeam == nil || profile.RosterTeam.ID != 38 || profile.RosterTeam.Name != "Roster Team" || profile.RosterTeam.JoinedAt != 200 || profile.RosterTeam.RetrievedAt != 300 {
		t.Fatalf("roster team: %+v %v", profile.RosterTeam, err)
	}
	profiles, err := repo.GetProfiles(t.Context(), []int{42, 43})
	if err != nil || len(profiles) != 1 || profiles[42].TeamID != 37 {
		t.Fatalf("profiles batch: %+v %v", profiles, err)
	}
	if _, err := repo.GetByID(t.Context(), 43); err == nil {
		t.Fatal("missing player must not resolve")
	}
	// Going private keeps the last public location; a later public check may clear it.
	if err := repo.SaveSteamEnrichment(t.Context(), 42, &model.Player{Name: "Steam name", SteamLocation: "Odesa", SteamPrivacy: "public"}, now, "ok"); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveSteamEnrichment(t.Context(), 42, &model.Player{Name: "Steam name", SteamPrivacy: "private"}, now.Add(time.Hour), "ok"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Query(t.Context(), `RETURN DOCUMENT("players", "42")`, nil, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.SteamLocation != "Odesa" || saved.SteamPrivacy != "private" || saved.SteamCheckedAt != now.Add(time.Hour).UnixMilli() || saved.SteamPublicAt != now.UnixMilli() {
		t.Fatalf("private profile lost last public data: %+v", saved)
	}
	ids, _, err = repo.GetSitemapPlayers(t.Context(), 0, 10)
	if err != nil || len(ids) != 1 || ids[0] != 42 {
		t.Fatalf("sitemap players: %v %v", ids, err)
	}
	if _, err := repo.SaveProfile(t.Context(), &model.Player{ID: 44, Name: "Steam only", ProfileSource: "steam"}); err != nil {
		t.Fatal(err)
	}
	if ids, total, err := repo.GetSitemapPlayers(t.Context(), 0, 10); err != nil || total != 1 || len(ids) != 1 {
		t.Fatalf("steam-only player listed in sitemap: %v %d %v", ids, total, err)
	}
	listed := true
	rows, total, err := repo.GetAll(t.Context(), 0, 10, model.PlayerFilter{Pro: &listed})
	if err != nil || total != 1 || len(rows) != 1 || rows[0].ID != 42 || rows[0].TeamName != "New Team" {
		t.Fatalf("pro directory: %+v %d %v", rows, total, err)
	}
	rows, total, err = repo.GetAll(t.Context(), 0, 10, model.PlayerFilter{Search: "44"})
	if err != nil || total != 1 || rows[0].ID != 44 {
		t.Fatalf("account search: %+v %d %v", rows, total, err)
	}
	rows, _, err = repo.GetAll(t.Context(), 0, 10, model.PlayerFilter{})
	if err != nil || len(rows) != 2 || rows[0].ID != 42 {
		t.Fatalf("listed pros should rank first: %+v %v", rows, err)
	}
}
