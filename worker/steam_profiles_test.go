package worker

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"testing"
	"testing/synctest"
	"time"
)

type steamRefreshRepo struct {
	PlayerRepository
	statuses []string
	stored   map[int]*model.Player
}

func (r *steamRefreshRepo) GetSteamRefreshCandidates(context.Context, time.Time, int) ([]int, error) {
	return []int{1, 2, 3}, nil
}
func (r *steamRefreshRepo) GetByID(_ context.Context, id int) (*model.Player, error) {
	if p, ok := r.stored[id]; ok {
		copied := *p
		return &copied, nil
	}
	return nil, &e.Error{Code: e.ENOTFOUND}
}

// SaveSteamEnrichment mirrors the repository rules: failures keep old data, a private
// profile keeps its last public location, and every attempt records a status.
func (r *steamRefreshRepo) SaveSteamEnrichment(_ context.Context, id int, profile *model.Player, _ time.Time, status string) error {
	r.statuses = append(r.statuses, status)
	if p, ok := r.stored[id]; ok {
		p.SteamStatus = status
		if profile != nil {
			p.SteamName, p.SteamPrivacy = profile.Name, profile.SteamPrivacy
			if profile.SteamLocation != "" || profile.SteamPrivacy == "public" {
				p.SteamLocation = profile.SteamLocation
			}
		}
	}
	return nil
}

func TestSteamSweepRecordsPrivacyChangesAndKeepsLastPublicData(t *testing.T) {
	store := &capturedUpdates{}
	repo := &steamRefreshRepo{stored: map[int]*model.Player{
		1: {ID: 1, Name: "Pro", SteamName: "old", SteamLocation: "Kyiv", SteamStatus: "ok", SteamPrivacy: "public"},
		2: {ID: 2, SteamName: "solo", SteamStatus: "ok", SteamPrivacy: "public"},
		3: {ID: 3, SteamName: "gone", SteamLocation: "Lviv", SteamStatus: "ok", SteamPrivacy: "public"},
	}}
	dl := &DataLoader{PlayerRepository: repo, Updates: store}
	err := dl.refreshSteamProfiles(t.Context(), func(_ context.Context, id int) (*model.Player, error) {
		switch id {
		case 1:
			return &model.Player{ID: 1, Name: "old", SteamPrivacy: "private"}, nil
		case 2:
			return &model.Player{ID: 2, Name: "solo", SteamPrivacy: "public"}, nil
		}
		return nil, &e.Error{Code: e.ENOTFOUND}
	})
	if err != nil || len(store.rows) != 2 {
		t.Fatalf("sweep: %v rows=%+v", err, store.rows)
	}
	private := store.rows[0]
	if private.EntityID != 1 || private.Name != "Pro" || private.URL != "/player/1" || len(private.Changes) != 1 ||
		private.Changes[0].Field != "steam_profile" || private.Changes[0].Before != "public" || private.Changes[0].After != "private" {
		t.Fatalf("privacy change not recorded as the only change: %+v", private)
	}
	if repo.stored[1].SteamLocation != "Kyiv" {
		t.Fatal("last public location dropped after the profile went private")
	}
	unavailable := store.rows[1]
	if unavailable.EntityID != 3 || unavailable.Name != "gone" || unavailable.Changes[0].After != "unavailable" || repo.stored[3].SteamName != "gone" {
		t.Fatalf("unavailable profile should keep data and log the state: %+v", unavailable)
	}
}

// Pacing belongs to the Steam rate limiter, so the sweep itself adds no delay.
func TestSteamSweepRecordsUnavailableProfilesWithoutPausing(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := &steamRefreshRepo{}
		dl := &DataLoader{PlayerRepository: repo}
		start := time.Now()
		err := dl.refreshSteamProfiles(t.Context(), func(_ context.Context, id int) (*model.Player, error) {
			if id == 2 {
				return nil, &e.Error{Code: e.ENOTFOUND}
			}
			return &model.Player{ID: id, Name: "Steam"}, nil
		})
		if err != nil || time.Since(start) != 0 || len(repo.statuses) != 3 || repo.statuses[1] != "unavailable" {
			t.Fatalf("sweep: %v %+v elapsed=%v", err, repo.statuses, time.Since(start))
		}
	})
}
func TestSteamSweepStopsOnTransportFailure(t *testing.T) {
	repo := &steamRefreshRepo{}
	dl := &DataLoader{PlayerRepository: repo}
	err := dl.refreshSteamProfiles(t.Context(), func(context.Context, int) (*model.Player, error) { return nil, context.DeadlineExceeded })
	if err != context.DeadlineExceeded || len(repo.statuses) != 1 || repo.statuses[0] != "request_failed" {
		t.Fatalf("unexpected result: %v %+v", err, repo.statuses)
	}
}
func TestSteamSweepCancellationDoesNotRecordFailure(t *testing.T) {
	repo := &steamRefreshRepo{}
	dl := &DataLoader{PlayerRepository: repo}
	ctx, cancel := context.WithCancel(t.Context())
	err := dl.refreshSteamProfiles(ctx, func(context.Context, int) (*model.Player, error) { cancel(); return nil, ctx.Err() })
	if err != context.Canceled || len(repo.statuses) != 0 {
		t.Fatalf("cancel recorded as failure: %v %+v", err, repo.statuses)
	}
}
