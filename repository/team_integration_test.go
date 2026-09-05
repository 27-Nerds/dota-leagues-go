package repository

import (
	"context"
	"dota_league/db"
	"dota_league/model"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestTeamsSortedByCompetitiveActivity(t *testing.T) {
	url := os.Getenv("ARANGO_TEST_URL")
	if url == "" {
		t.Skip("set ARANGO_TEST_URL to run against a local ArangoDB instance")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	conn, err := db.Connect(ctx, url, "", "", fmt.Sprintf("team_sort_%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := conn.DB.Remove(ctx); err != nil {
			t.Errorf("remove temporary database: %v", err)
		}
	})
	now := time.Now().Unix()
	repo := NewTeamRepository(conn)
	for id := 0; id <= 9; id++ {
		// Even the dormant teams have fresh metadata and professional status.
		country := "US"
		if id == 1 || id == 2 || id == 6 {
			country = "ua"
		}
		if err := repo.Store(ctx, &model.Team{ID: id, Name: fmt.Sprintf("Team %d", id), CountryCode: country, Wins: id, UpdatedTimestamp: now, Pro: true}); err != nil {
			t.Fatal(err)
		}
	}
	for _, league := range []model.LeagueDetails{{ID: 1, Tier: 2}, {ID: 2, Tier: 1}} {
		if err := NewLeagueDetailsRepository(conn).Store(ctx, &league); err != nil {
			t.Fatal(err)
		}
	}
	series := NewLeagueSeriesRepository(conn)
	if err := series.ReplaceAllForLeague(ctx, 1, []model.SeriesInfo{
		{SeriesID: 1, TeamID1: 1, StartTime: now - 180*86400, MatchIDs: []string{"1"}},
		{SeriesID: 2, TeamID1: 3, StartTime: now - 7*86400, MatchIDs: []string{"2"}},
		{SeriesID: 3, TeamID1: 5, TeamID2: 7, StartTime: now - 86400, MatchIDs: []string{"3"}},
		// A future timestamp or an unplayed series must not count as activity.
		{SeriesID: 4, TeamID1: 8, StartTime: now + 86400, MatchIDs: []string{"4"}},
		{SeriesID: 5, TeamID1: 9, StartTime: now - 60},
	}); err != nil {
		t.Fatal(err)
	}
	if err := series.ReplaceAllForLeague(ctx, 2, []model.SeriesInfo{
		{SeriesID: 6, TeamID2: 4, StartTime: now - 60, MatchIDs: []string{"6"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := NewGameRepository(conn).StoreAll(ctx, []model.Game{{ServerSteamID: "live", DireTeamID: 2}}); err != nil {
		t.Fatal(err)
	}
	check := func(offset, limit int, want []int) {
		t.Helper()
		teams, total, err := repo.GetAll(ctx, offset, limit, model.TeamFilter{})
		if err != nil {
			t.Fatal(err)
		}
		var ids []int
		for _, team := range teams {
			ids = append(ids, team.ID)
		}
		if total != 9 || !reflect.DeepEqual(ids, want) {
			t.Fatalf("offset=%d limit=%d: IDs=%v total=%d, want %v total=9", offset, limit, ids, total, want)
		}
	}
	check(0, 20, []int{2, 5, 7, 3, 4, 1, 6, 8, 9})
	check(0, 3, []int{2, 5, 7})
	check(3, 3, []int{3, 4, 1})
	if err := conn.EnsureUpdatesCollection(ctx); err != nil {
		t.Fatal(err)
	}
	for _, update := range []model.Update{
		{Entity: "roster", EntityID: 6, Action: "updated", CreatedAt: (now - 30) * 1000},
		{Entity: "team", EntityID: 8, Action: "created", CreatedAt: (now - 30) * 1000},
		{Entity: "team", EntityID: 9, Action: "updated", CreatedAt: (now + 86400) * 1000},
		{Entity: "team", EntityID: 1, Action: "updated", CreatedAt: (now - 180*86400) * 1000},
	} {
		if err := conn.Insert(ctx, "updates", update); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		filter        model.TeamFilter
		offset, limit int
		ids           []int
		total         int64
	}{
		{model.TeamFilter{Country: " UA ", Sort: "name"}, 0, 100, []int{1, 2, 6}, 3},
		{model.TeamFilter{Active: true, Sort: "name"}, 0, 100, []int{2, 3, 4, 5, 6, 7}, 6},
		{model.TeamFilter{Active: true, ActiveDays: 2, Sort: "name"}, 0, 100, []int{2, 4, 5, 6, 7}, 5},
		{model.TeamFilter{Country: "ua", Active: true, ActiveDays: 2, Sort: "wins"}, 1, 1, []int{2}, 2},
		{model.TeamFilter{Sort: "activity"}, 0, 3, []int{2, 6, 4}, 9},
		{model.TeamFilter{Sort: "wins", Order: "asc"}, 0, 3, []int{1, 2, 3}, 9},
	} {
		rows, total, err := repo.GetAll(ctx, tc.offset, tc.limit, tc.filter)
		ids := []int{}
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		if err != nil || total != tc.total || !reflect.DeepEqual(ids, tc.ids) {
			t.Fatalf("filter %+v: ids=%v total=%d err=%v", tc.filter, ids, total, err)
		}
	}
	if err := repo.Update(ctx, &model.Team{ID: 8, Name: "Hidden %_TEAM", Tag: "Secret"}); err != nil {
		t.Fatal(err)
	}
	for _, professional := range []bool{true, false} {
		rows, total, err := repo.GetAll(ctx, 0, 100, model.TeamFilter{Pro: &professional})
		want := int64(8)
		if !professional {
			want = 1
		}
		if err != nil || total != want || int64(len(rows)) != want {
			t.Fatalf("pro=%v: rows=%d total=%d err=%v", professional, len(rows), total, err)
		}
		for _, row := range rows {
			if row.Pro != professional {
				t.Fatalf("pro=%v returned team %d with pro=%v", professional, row.ID, row.Pro)
			}
		}
	}
	for _, search := range []string{" hidden ", "SECRET", "%_"} {
		teams, total, err := repo.GetAll(ctx, 0, 1, model.TeamFilter{Search: search})
		if err != nil || total != 1 || len(teams) != 1 || teams[0].ID != 8 {
			t.Fatalf("search %q must find a team beyond the first page: teams=%v total=%d err=%v", search, teams, total, err)
		}
	}
	teams, total, err := repo.GetAll(ctx, 1, 1, model.TeamFilter{Search: "team"})
	if err != nil || total != 9 || len(teams) != 1 || teams[0].ID != 5 {
		t.Fatalf("search pagination must retain ranking and filtered total: teams=%v total=%d err=%v", teams, total, err)
	}
	teams, total, err = repo.GetAll(ctx, 0, 100, model.TeamFilter{Search: "no such team"})
	if err != nil || total != 0 || teams == nil || len(teams) != 0 {
		t.Fatalf("empty search must return an empty array: teams=%v total=%d err=%v", teams, total, err)
	}
	if err := NewGameRepository(conn).RemoveAll(ctx); err != nil {
		t.Fatal(err)
	}
	check(0, 3, []int{5, 7, 3})

	// Profiles resolve names without inventing pages for absent league details.
	if err := NewPlayerRepository(conn).Store(ctx, &model.Player{ID: 123, Name: "Stored player"}); err != nil {
		t.Fatal(err)
	}
	if err := NewLeagueDetailsRepository(conn).Store(ctx, &model.LeagueDetails{ID: 100, Name: "Loaded tournament"}); err != nil {
		t.Fatal(err)
	}
	if err := NewLeagueRepository(conn).Store(ctx, &model.League{ID: 101, Name: "Catalogue tournament"}); err != nil {
		t.Fatal(err)
	}
	var profile model.Team
	if err := json.Unmarshal([]byte(`{"team_id":12,"members":[{"account_id":123,"pro_name":""},{"account_id":999,"pro_name":"Roster name"}],"dpc_results":[{"league_id":100},{"league_id":101},{"league_id":999}]}`), &profile); err != nil {
		t.Fatal(err)
	}
	if err := repo.Store(ctx, &profile); err != nil {
		t.Fatal(err)
	}
	resolved, err := repo.GetByID(ctx, 12)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Members[0].PlayerName != "Stored player" || resolved.Members[1].ProName != "Roster name" {
		t.Fatalf("unresolved members: %+v", resolved.Members)
	}
	if resolved.DpcResults[0].LeagueName != "Loaded tournament" || !resolved.DpcResults[0].LeagueAvailable || resolved.DpcResults[1].LeagueName != "Catalogue tournament" || resolved.DpcResults[1].LeagueAvailable || resolved.DpcResults[2].LeagueAvailable {
		t.Fatalf("incorrect history links: %+v", resolved.DpcResults)
	}
}
