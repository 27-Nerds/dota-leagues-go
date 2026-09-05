package worker

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"testing"
	"testing/synctest"
	"time"
)

type existingPlayerQueue struct {
	PlayerRepository
	checks int
}

func (p *existingPlayerQueue) NeedsProfileRefresh(context.Context, int, time.Time) (bool, error) {
	p.checks++
	return false, nil
}
func TestQueuedExistingPlayerSkipsFetchAndInsert(t *testing.T) {
	p := &existingPlayerQueue{}
	loader := &DataLoader{PlayerRepository: p}
	if err := loader.storeSinglePlayer(t.Context(), 435600912); err != nil || p.checks != 1 {
		t.Fatal("queued duplicate not skipped", err)
	}
}

type finishedGamesRepo struct {
	GameRepository
	cleared bool
}

func (r *finishedGamesRepo) GetAll(context.Context) ([]model.Game, error) {
	return []model.Game{{LeagueID: 99}}, nil
}
func (r *finishedGamesRepo) RemoveAll(context.Context) error { r.cleared = true; return nil }

type missingFinishedLeague struct{ LeagueDetailsRepository }

func (missingFinishedLeague) UpdateLiveStatus(context.Context, int, bool) error {
	return &e.Error{Code: e.ENOTFOUND}
}
func TestMissingFinishedLeagueDoesNotBlockGameCleanup(t *testing.T) {
	games := &finishedGamesRepo{}
	manager := NewLiveGamesManager(t.Context(), nil)
	defer manager.Stop()
	loader := &DataLoader{GameRepository: games, LeagueDetailsRepository: missingFinishedLeague{}, LiveGamesManager: manager}
	if err := loader.applyGamesUpdate(t.Context(), nil); err != nil {
		t.Fatal(err)
	}
	if !games.cleared {
		t.Fatal("stale games not cleared")
	}
}
func TestLivePollingAdaptsAndDisappearedServersStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager := NewLiveGamesManager(t.Context(), nil)
		defer manager.Stop()
		manager.load = func(ctx context.Context, _ string) (*model.LiveGameDetails, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		manager.AddGame(model.Game{ServerSteamID: "one"})
		for _, id := range []string{"two", "three", "four", "five", "six"} {
			manager.AddGame(model.Game{ServerSteamID: id})
		}
		if manager.pollInterval() < 8*time.Second {
			t.Fatal("polling exceeds half the default request budget")
		}
		synctest.Wait()
		manager.ReplaceGames(nil)
		synctest.Wait()
		manager.mu.Lock()
		remaining := len(manager.liveGames)
		manager.mu.Unlock()
		if remaining != 0 {
			t.Fatal("removed servers still polling")
		}
	})
}
func TestUnavailableLiveStatsBackOff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager := NewLiveGamesManager(t.Context(), nil)
		defer manager.Stop()
		calls := 0
		manager.load = func(context.Context, string) (*model.LiveGameDetails, error) {
			calls++
			return nil, &e.Error{Code: e.ENOTFOUND}
		}
		manager.AddGame(model.Game{ServerSteamID: "unavailable"})
		time.Sleep(61 * time.Second)
		synctest.Wait()
		if calls != 3 {
			t.Fatalf("want attempts at 0,15,45 seconds; got %d", calls)
		}
	})
}

type missingPlayerQueue struct {
	PlayerRepository
	stored bool
}

func (p *missingPlayerQueue) NeedsProfileRefresh(context.Context, int, time.Time) (bool, error) {
	return !p.stored, nil
}
func (p *missingPlayerQueue) SaveProfile(context.Context, *model.Player) (bool, error) {
	p.stored = true
	return true, nil
}

func TestMissingPlayerCooldownAndRecovery(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		repo := &missingPlayerQueue{}
		dl := &DataLoader{PlayerRepository: repo, Updates: &capturedUpdates{}}
		calls := 0
		load := func(context.Context, int) (*model.Player, error) {
			calls++
			if calls == 1 {
				return nil, &e.Error{Code: e.ENOTFOUND}
			}
			return &model.Player{ID: 293510753}, nil
		}
		for range 3 {
			if err := dl.storeSinglePlayerWithLoader(t.Context(), 293510753, load); err != nil {
				t.Fatal(err)
			}
		}
		if calls != 1 || repo.stored {
			t.Fatalf("missing profile fetched %d times, stored=%v", calls, repo.stored)
		}
		time.Sleep(missingPlayerRetryInterval)
		if err := dl.storeSinglePlayerWithLoader(t.Context(), 293510753, load); err != nil {
			t.Fatal(err)
		}
		if calls != 2 || !repo.stored {
			t.Fatal("profile did not recover after cooldown")
		}
	})
}

func TestPlayerTransientFailureIsNotCached(t *testing.T) {
	dl := &DataLoader{PlayerRepository: &missingPlayerQueue{}}
	calls := 0
	load := func(context.Context, int) (*model.Player, error) { calls++; return nil, context.DeadlineExceeded }
	for range 2 {
		if err := dl.storeSinglePlayerWithLoader(t.Context(), 42, load); err != context.DeadlineExceeded {
			t.Fatal(err)
		}
	}
	if calls != 2 || len(dl.missingPlayers) != 0 {
		t.Fatal("transient error cached as missing profile")
	}
}

func TestMissingPlayerCacheBoundedAndExpiredEntriesRemoved(t *testing.T) {
	dl := &DataLoader{}
	for id := range missingPlayerCacheLimit + 1 {
		dl.rememberMissingPlayer(id)
	}
	if len(dl.missingPlayers) != missingPlayerCacheLimit {
		t.Fatal("cache exceeds limit")
	}
	dl.missingPlayers[99] = time.Now().Add(-time.Second)
	dl.rememberMissingPlayer(missingPlayerCacheLimit + 2)
	if _, ok := dl.missingPlayers[99]; ok {
		t.Fatal("expired entry retained")
	}
}
