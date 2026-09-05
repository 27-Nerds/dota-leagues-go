// Package worker refreshes league, team, player, and live-match data.
package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// channelBuffer absorbs short bursts; producers wait when consumers fall behind.
const channelBuffer = 1024
const detailsRefreshInterval = 2 * time.Hour

// DataLoader owns the background refresh jobs and their shutdown.
type DataLoader struct {
	LeagueRepository          LeagueRepository
	LeagueDetailsRepository   LeagueDetailsRepository
	GameRepository            GameRepository
	PlayerRepository          PlayerRepository
	TeamRepository            TeamRepository
	TeamRosterRepository      TeamRosterRepository
	LeagueSeriesRepository    LeagueSeriesRepository
	LiveGameDetailsRepository LiveGameDetailsRepository
	LoadLeagueDetails         chan int
	LoadTeam                  chan int
	LoadSinglePlayer          chan int
	LiveGamesManager          *LiveGamesManager
	Updates                   UpdatesStore
	// missingPlayers is owned by the single player queue consumer.
	missingPlayers map[int]time.Time
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewDataLoader starts refresh jobs that live until ctx is canceled or Stop is called.
func NewDataLoader(
	ctx context.Context,
	lr LeagueRepository,
	ldr LeagueDetailsRepository,
	gr GameRepository,
	pr PlayerRepository,
	tr TeamRepository,
	trr TeamRosterRepository,
	lsr LeagueSeriesRepository,
	lgdr LiveGameDetailsRepository,
	updates UpdatesStore,
) *DataLoader {
	ctx, cancel := context.WithCancel(ctx)
	dl := &DataLoader{
		LeagueRepository: lr, LeagueDetailsRepository: ldr,
		GameRepository: gr, PlayerRepository: pr,
		TeamRepository: tr, TeamRosterRepository: trr,
		LeagueSeriesRepository: lsr, LiveGameDetailsRepository: lgdr,
		LoadLeagueDetails: make(chan int, channelBuffer),
		LoadTeam:          make(chan int, channelBuffer),
		LoadSinglePlayer:  make(chan int, channelBuffer),
		cancel:            cancel,
		Updates:           updates,
	}
	dl.LiveGamesManager = NewLiveGamesManager(ctx, lgdr)
	dl.wg.Go(func() { consume(ctx, "league details", dl.LoadLeagueDetails, dl.storeLeagueDetails) })
	dl.wg.Go(func() { consume(ctx, "team", dl.LoadTeam, dl.storeTeam) })
	dl.wg.Go(func() { consume(ctx, "player", dl.LoadSinglePlayer, dl.storeSinglePlayer) })
	dl.wg.Go(func() {
		// Reset stale live flags before starting refreshes.
		reportUpdate(ctx, "reset league live status", dl.LeagueDetailsRepository.SetAllAsNotLive(ctx))
		if ctx.Err() != nil {
			return
		}
		dl.wg.Go(func() { runPeriodic(ctx, "leagues", 2*time.Second, 2*time.Hour, dl.performLeaguesUpdate) })
		dl.wg.Go(func() {
			runPeriodic(ctx, "historical leagues", 30*time.Second, 10*time.Minute, dl.performHistoricalLeaguesUpdate)
		})
		dl.wg.Go(func() { runPeriodic(ctx, "teams", 20*time.Second, 10*time.Minute, dl.performTeamsUpdate) })
		dl.wg.Go(func() { runPeriodic(ctx, "games", time.Minute, time.Minute, dl.performGamesUpdate) })
		dl.wg.Go(func() { runPeriodic(ctx, "prizepool", time.Hour, time.Hour, dl.performPrizePoolUpdate) })
		dl.wg.Go(func() { runPeriodic(ctx, "steam profiles", 45*time.Second, time.Minute, dl.performSteamProfilesUpdate) })
		dl.wg.Go(func() { runPeriodic(ctx, "players", 10*time.Second, 2*time.Hour, dl.performPlayersUpdate) })
	})
	return dl
}

// Stop cancels in-flight work and waits for every worker to exit. It is safe to call repeatedly.
func (dl *DataLoader) Stop() {
	dl.cancel()
	dl.wg.Wait()
	dl.LiveGamesManager.Stop()
}

// Each schedule owns one timer and executes one update at a time. The next delay
// starts after completion, so slow upstream responses cannot pile up refresh jobs.
func runPeriodic(ctx context.Context, name string, first, interval time.Duration, update func(context.Context) error) {
	timer := time.NewTimer(first)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if ctx.Err() != nil {
				return
			}
			reportUpdate(ctx, name, update(ctx))
			timer.Reset(interval)
		}
	}
}

func consume(ctx context.Context, name string, jobs <-chan int, update func(context.Context, int) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case id, ok := <-jobs:
			if !ok || ctx.Err() != nil {
				return
			}
			reportUpdate(ctx, name, update(ctx, id))
		}
	}
}

// Queues remain open: cancellation releases both consumers and blocked producers.
func enqueue(ctx context.Context, jobs chan<- int, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case jobs <- id:
		return nil
	}
}

func reportUpdate(ctx context.Context, name string, err error) {
	if err != nil && ctx.Err() == nil {
		slog.ErrorContext(ctx, "refresh failed", "job", name, "error", err)
	}
}
