package worker

import (
	"context"
	"dota_league/model"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestPeriodicUpdateAndBlockedQueueStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		var calls atomic.Int32
		go runPeriodic(ctx, "test", time.Second, time.Second, func(ctx context.Context) error {
			calls.Add(1)
			<-ctx.Done()
			return ctx.Err()
		})
		queueResult := make(chan error, 1)
		go func() { queueResult <- enqueue(ctx, make(chan int), 1) }()
		time.Sleep(10 * time.Second)
		synctest.Wait()
		if calls.Load() != 1 {
			t.Fatalf("slow update overlapped: %d calls", calls.Load())
		}
		cancel()
		synctest.Wait()
		if err := <-queueResult; !errors.Is(err, context.Canceled) {
			t.Fatalf("blocked producer did not stop: %v", err)
		}
	})
}

func TestLiveGamesConcurrentAddAndStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager := NewLiveGamesManager(t.Context(), nil)
		var calls atomic.Int32
		manager.load = func(ctx context.Context, id string) (*model.LiveGameDetails, error) {
			calls.Add(1)
			<-ctx.Done()
			return nil, ctx.Err()
		}
		var adders sync.WaitGroup
		for range 50 {
			adders.Go(func() { manager.AddGame(model.Game{ServerSteamID: "one"}) })
		}
		adders.Wait()
		synctest.Wait()
		if calls.Load() != 1 {
			t.Fatalf("duplicate game workers: %d", calls.Load())
		}
		var stoppers sync.WaitGroup
		stoppers.Go(manager.Stop)
		for range 50 {
			stoppers.Go(func() { manager.AddGame(model.Game{ServerSteamID: "two"}) })
		}
		stoppers.Wait()
		manager.Stop()
		manager.AddGame(model.Game{ServerSteamID: "after-stop"})
		if len(manager.liveGames) != 0 {
			t.Fatal("live workers remained after shutdown")
		}
	})
}

func TestStalledLiveGameExpires(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		manager := NewLiveGamesManager(t.Context(), nil)
		defer manager.Stop()
		manager.load = func(ctx context.Context, id string) (*model.LiveGameDetails, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		manager.AddGame(model.Game{ServerSteamID: "stalled"})
		time.Sleep(liveGameTimeout + time.Second)
		synctest.Wait()
		manager.mu.Lock()
		remaining := len(manager.liveGames)
		manager.mu.Unlock()
		if remaining != 0 {
			t.Fatal("stalled request kept its game alive past the inactivity deadline")
		}
	})
}
