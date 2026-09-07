package worker

import (
	"context"
	"dota_league/model"
	"errors"
	"testing"
	"time"
)

type capturedUpdates struct{ rows []model.Update }

func (s *capturedUpdates) Store(_ context.Context, row *model.Update) error {
	s.rows = append(s.rows, *row)
	return nil
}

func TestUpdatesIgnoreRefreshMetadata(t *testing.T) {
	store := &capturedUpdates{}
	dl := &DataLoader{Updates: store}
	team := model.Team{ID: 7, Name: "Team A", UpdatedTimestamp: 1}
	dl.recordUpdate(t.Context(), "team", team.ID, team.Name, nil, teamSnapshot(&team))
	before := teamSnapshot(&team)
	team.UpdatedTimestamp = 2
	team.DBKey = "7"
	dl.recordUpdate(t.Context(), "team", team.ID, team.Name, before, teamSnapshot(&team))
	team.Name = "Team B"
	dl.recordUpdate(t.Context(), "team", team.ID, team.Name, before, teamSnapshot(&team))
	if len(store.rows) != 2 || store.rows[0].Action != "created" {
		t.Fatalf("unexpected events: %+v", store.rows)
	}
	row := store.rows[1]
	if row.Action != "updated" || row.URL != "/team/7" || len(row.Changes) != 1 ||
		row.Changes[0].Field != "name" || row.Changes[0].Before != "Team A" || row.Changes[0].After != "Team B" {
		t.Fatalf("incorrect public change: %+v", row)
	}
}

func TestPlayerProfileChangesRecordTeamMovesAndIdentity(t *testing.T) {
	store := &capturedUpdates{}
	dl := &DataLoader{Updates: store}
	before := &model.Player{ID: 5, Name: "Fallback", ProfileSource: "steam", SteamName: "Fallback", SteamNextRefreshAt: 1}
	after := &model.Player{ID: 5, Name: "Pro", TeamID: 7, TeamName: "Team A", IsPro: true, TotalEarnings: 100, SteamNextRefreshAt: 2}
	dl.recordUpdate(t.Context(), "player", after.ID, after.Name, playerSnapshot(before), playerSnapshot(after))
	moved := *after
	moved.TeamID, moved.TeamName = 8, ""
	dl.recordUpdate(t.Context(), "player", moved.ID, moved.Name, playerSnapshot(after), playerSnapshot(&moved))
	dl.recordUpdate(t.Context(), "player", moved.ID, moved.Name, playerSnapshot(&moved), playerSnapshot(&moved))
	if len(store.rows) != 2 || store.rows[0].URL != "/player/5" {
		t.Fatalf("unexpected events: %+v", store.rows)
	}
	fields := map[string]model.UpdateChange{}
	for _, c := range store.rows[0].Changes {
		fields[c.Field] = c
	}
	if fields["name"].After != "Pro" || fields["team"].After != "Team A" || fields["is_pro"].After != true || fields["total_earnings"].After != 100 {
		t.Fatalf("identity changes missing: %+v", store.rows[0].Changes)
	}
	if _, refresh := fields["steam_next_refresh_at"]; refresh {
		t.Fatal("refresh bookkeeping leaked into the feed")
	}
	if store.rows[1].Changes[0].Field != "team" || store.rows[1].Changes[0].After != "Team #8" {
		t.Fatalf("team move without a stored name should fall back to the ID: %+v", store.rows[1].Changes)
	}
}

type rosterStore struct {
	TeamRosterRepository
	stored *model.TeamRoster
	err    error
	writes int
}

func (s *rosterStore) ExistsByTeamID(context.Context, int) (bool, error)       { return s.stored != nil, nil }
func (s *rosterStore) GetByID(context.Context, int) (*model.TeamRoster, error) { return s.stored, nil }
func (s *rosterStore) Update(context.Context, *model.TeamRoster) error {
	s.writes++
	return s.err
}
func (s *rosterStore) Store(context.Context, *model.TeamRoster) error {
	s.writes++
	return s.err
}

type existingPlayers struct{ PlayerRepository }

func (existingPlayers) NeedsProfileRefresh(context.Context, int, time.Time) (bool, error) {
	return false, nil
}

func TestRosterUpdatesFollowSuccessfulWrites(t *testing.T) {
	feed := &capturedUpdates{}
	repo := &rosterStore{stored: &model.TeamRoster{TeamID: 7, TeamMembers: []model.TeamMember{
		{AccountID: 1, IsActive: true}, {AccountID: 2, IsActive: true},
	}}}
	dl := &DataLoader{Updates: feed, TeamRosterRepository: repo, PlayerRepository: existingPlayers{}}
	team := &model.Team{ID: 7, Name: "Team A"}
	// The upstream member order can change without a roster transfer.
	for _, id := range []int{2, 1} {
		team.Members = append(team.Members, struct {
			AccountID     int    `json:"account_id"`
			TimeJoined    int    `json:"time_joined"`
			Admin         bool   `json:"admin"`
			ProName       string `json:"pro_name"`
			AvatarURL     string `json:"avatar_url,omitempty"`
			SteamLocation string `json:"steam_location,omitempty"`
			PlayerName    string `json:"player_name,omitempty"`
			RealName      string `json:"real_name,omitempty"`
			Role          int    `json:"role"`
		}{AccountID: id})
	}
	if err := dl.storeTeamRoster(t.Context(), team, true); err != nil {
		t.Fatal(err)
	}
	if len(feed.rows) != 0 {
		t.Fatal("member reordering created an event")
	}
	team.Members[0].AccountID = 3
	repo.err = errors.New("write failed")
	if err := dl.storeTeamRoster(t.Context(), team, true); err == nil {
		t.Fatal("expected write failure")
	}
	if len(feed.rows) != 0 {
		t.Fatal("failed write created an event")
	}
	repo.err = nil
	if err := dl.storeTeamRoster(t.Context(), team, true); err != nil {
		t.Fatal(err)
	}
	if len(feed.rows) != 1 || feed.rows[0].Entity != "roster" || feed.rows[0].Changes[0].Field != "members" {
		t.Fatalf("missing roster change: %+v", feed.rows)
	}
}

func TestInitialRosterDoesNotCreateSeparateActivity(t *testing.T) {
	for _, tc := range []struct {
		name       string
		teamExists bool
		stored     *model.TeamRoster
	}{
		{name: "new team"},
		{name: "first roster for existing team", teamExists: true},
		{name: "retry after roster saved but team creation failed", stored: &model.TeamRoster{
			TeamID: 7, TeamMembers: []model.TeamMember{{AccountID: 1, IsActive: true}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			feed := &capturedUpdates{}
			repo := &rosterStore{stored: tc.stored}
			dl := &DataLoader{Updates: feed, TeamRosterRepository: repo, PlayerRepository: existingPlayers{}}
			if err := dl.storeTeamRoster(t.Context(), &model.Team{ID: 7, Name: "Team A"}, tc.teamExists); err != nil {
				t.Fatal(err)
			}
			if repo.writes != 1 || len(feed.rows) != 0 {
				t.Fatalf("expected saved roster without activity: writes=%d events=%+v", repo.writes, feed.rows)
			}
		})
	}
}

type updateTeamLookup struct {
	TeamRepository
	team *model.Team
	err  error
}

func (s updateTeamLookup) GetByID(context.Context, int) (*model.Team, error) { return s.team, s.err }

func TestPlayerUpdatesCaptureTeam(t *testing.T) {
	for _, tc := range []struct {
		name   string
		player model.Player
		lookup updateTeamLookup
		want   model.UpdateTeam
	}{
		{name: "API team", player: model.Player{ID: 1, TeamID: 7, TeamName: "Team A"}, want: model.UpdateTeam{ID: 7, Name: "Team A"}},
		{name: "stored team name", player: model.Player{ID: 1, TeamID: 7}, lookup: updateTeamLookup{team: &model.Team{ID: 7, Name: "Team A"}}, want: model.UpdateTeam{ID: 7, Name: "Team A"}},
		{name: "team lookup unavailable", player: model.Player{ID: 1, TeamID: 7}, lookup: updateTeamLookup{err: errors.New("offline")}, want: model.UpdateTeam{ID: 7}},
		{name: "no team", player: model.Player{ID: 1}, want: model.UpdateTeam{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			feed := &capturedUpdates{}
			dl := &DataLoader{Updates: feed, TeamRepository: tc.lookup}
			dl.recordPlayerUpdate(t.Context(), &tc.player)
			tc.player.TeamName = "Changed later"
			if len(feed.rows) != 1 {
				t.Fatalf("events: %+v", feed.rows)
			}
			row := feed.rows[0]
			if row.Team == nil || *row.Team != tc.want || row.Action != "created" || row.Entity != "player" || len(row.Changes) != 0 {
				t.Fatalf("unexpected event: %+v", row)
			}
		})
	}
}
