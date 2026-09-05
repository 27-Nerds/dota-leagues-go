package repository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

type activityDB struct {
	fakeDB
	query func(context.Context, any) error
}

func (f *activityDB) Query(ctx context.Context, _ string, _ map[string]any, out any) (string, error) {
	return "", f.query(ctx, out)
}

func TestCompetitiveActivityCacheExpiryAndFailure(t *testing.T) {
	now := time.Now()
	calls := 0
	fail := false
	conn := &activityDB{query: func(_ context.Context, out any) error {
		calls++
		if fail {
			return errors.New("database unavailable")
		}
		*out.(*map[string]teamActivity) = map[string]teamActivity{"1": {LastPlayed: int64(calls), Tier: 2}}
		return nil
	}}
	repo := NewTeamRepository(conn)
	first, err := repo.competitiveActivity(t.Context(), now)
	if err != nil {
		t.Fatal(err)
	}
	cached, err := repo.competitiveActivity(t.Context(), now.Add(teamActivityTTL-time.Nanosecond))
	if err != nil || calls != 1 || cached["1"] != first["1"] {
		t.Fatalf("cache missed: %v calls=%d", err, calls)
	}
	fail = true
	if _, err := repo.competitiveActivity(t.Context(), now.Add(teamActivityTTL)); err == nil {
		t.Fatal("refresh failure hidden")
	}
	fail = false
	refreshed, err := repo.competitiveActivity(t.Context(), now.Add(teamActivityTTL))
	if err != nil || calls != 3 || refreshed["1"].LastPlayed != 3 {
		t.Fatalf("failed refresh was cached: %v calls=%d", err, calls)
	}
}

func TestCompetitiveActivityConcurrentRefreshAndCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		gate := make(chan struct{})
		conn := &activityDB{query: func(ctx context.Context, out any) error {
			calls.Add(1)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-gate:
			}
			*out.(*map[string]teamActivity) = map[string]teamActivity{"1": {LastPlayed: 123}}
			return nil
		}}
		repo := NewTeamRepository(conn)
		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				result, err := repo.competitiveActivity(t.Context(), time.Now())
				if err != nil || result["1"].LastPlayed != 123 {
					t.Errorf("refresh: %v %v", result, err)
				}
			})
		}
		synctest.Wait()
		ctx, cancel := context.WithCancel(t.Context())
		canceled := make(chan error, 1)
		go func() { _, err := repo.competitiveActivity(ctx, time.Now()); canceled <- err }()
		synctest.Wait()
		cancel()
		synctest.Wait()
		if err := <-canceled; !errors.Is(err, context.Canceled) {
			t.Fatalf("waiter cancellation: %v", err)
		}
		if calls.Load() != 1 {
			t.Fatalf("concurrent refreshes: %d", calls.Load())
		}
		close(gate)
		wg.Wait()
	})
}
