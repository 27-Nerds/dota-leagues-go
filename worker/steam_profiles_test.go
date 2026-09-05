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
}

func (r *steamRefreshRepo) GetSteamRefreshCandidates(context.Context, time.Time, int) ([]int, error) {
	return []int{1, 2, 3}, nil
}
func (r *steamRefreshRepo) SaveSteamEnrichment(_ context.Context, _ int, _ *model.Player, _ time.Time, status string) error {
	r.statuses = append(r.statuses, status)
	return nil
}

func TestSteamSweepPacingAndUnavailableProfiles(t *testing.T) {
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
		if err != nil || time.Since(start) != 10*time.Second || len(repo.statuses) != 3 || repo.statuses[1] != "unavailable" {
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
