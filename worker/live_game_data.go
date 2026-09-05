package worker

import (
	"context"
	"dota_league/api"
	e "dota_league/error"
	"dota_league/model"
	"log/slog"
	"sync"
	"time"
)

const liveGameTimeout = 200 * time.Second

// LiveGamesManager owns one polling worker per server. The mutex protects only
// membership and shutdown; API and database calls always happen outside the lock.
type LiveGamesManager struct {
	mu        sync.Mutex
	liveGames map[string]context.CancelFunc
	stopped   bool
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	repo      LiveGameDetailsRepository
	load      func(context.Context, string) (*model.LiveGameDetails, error)
}

func NewLiveGamesManager(ctx context.Context, repo LiveGameDetailsRepository) *LiveGamesManager {
	ctx, cancel := context.WithCancel(ctx)
	return &LiveGamesManager{
		ctx: ctx, cancel: cancel, repo: repo,
		liveGames: make(map[string]context.CancelFunc),
		load:      api.GetLiveGameStats,
	}
}

func (m *LiveGamesManager) AddGame(game model.Game) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped || m.ctx.Err() != nil {
		return
	}
	id := game.ServerSteamID
	if _, exists := m.liveGames[id]; exists {
		return
	}
	if id == "" || id == "0" {
		return
	}
	pollCtx, cancel := context.WithCancel(m.ctx)
	m.liveGames[id] = cancel
	delay := time.Duration(float64(len(m.liveGames)-1) / liveRequestBudget() * float64(time.Second))
	// Register the task while holding the same lock Stop uses before waiting.
	m.wg.Go(func() { m.poll(pollCtx, id, delay) })
}

// Stop prevents new games, cancels in-flight requests, and joins all polling workers.
func (m *LiveGamesManager) Stop() {
	m.mu.Lock()
	m.stopped = true
	m.cancel()
	m.mu.Unlock()
	m.wg.Wait()
}

func (m *LiveGamesManager) poll(pollCtx context.Context, id string, initialDelay time.Duration) {
	defer func() {
		m.mu.Lock()
		m.liveGames[id]()
		delete(m.liveGames, id)
		m.mu.Unlock()
	}()
	deadline := time.Now().Add(initialDelay + liveGameTimeout)
	failures := 0
	timer := time.NewTimer(initialDelay)
	defer timer.Stop()
	for {
		select {
		case <-pollCtx.Done():
			return
		case <-timer.C:
		}
		if pollCtx.Err() != nil || !time.Now().Before(deadline) {
			return
		}
		// The inactivity deadline also bounds a stalled request or database write.
		ctx, cancel := context.WithDeadline(pollCtx, deadline)
		details, err := m.load(ctx, id)
		if err == nil {
			err = m.store(ctx, details)
		}
		cancel()
		if pollCtx.Err() != nil {
			return
		}
		if !time.Now().Before(deadline) {
			return
		}
		interval := m.pollInterval()
		if err == nil {
			failures = 0
			deadline = time.Now().Add(max(liveGameTimeout, 2*interval))
		} else {
			failures++
			// An ended/unavailable server is expected. Genuine failures remain
			// visible, with retries spaced out instead of a three-second loop.
			if !e.IsNotFound(err) && (failures == 1 || failures%4 == 0) {
				slog.WarnContext(m.ctx, "live game refresh failed", "server_id", id, "error", err)
			}
			backoff := 15 * time.Second * time.Duration(1<<min(failures-1, 2))
			interval = max(interval, backoff)
		}
		timer.Reset(min(interval, max(0, time.Until(deadline))))
	}
}

func (m *LiveGamesManager) store(ctx context.Context, details *model.LiveGameDetails) error {
	details.DBKey = details.Match.Matchid
	exists, err := m.repo.ExistsByID(ctx, details.Match.Matchid)
	if err != nil {
		return err
	}
	if exists {
		return m.repo.Update(ctx, details)
	}
	return m.repo.Store(ctx, details)
}

// Reserve roughly half of the shared budget for live stats; metadata, discovery
// and interactive requests also need capacity. Initial requests are staggered.
func liveRequestBudget() float64 { return max(0.01, api.ValveRateLimit()/2) }
func (m *LiveGamesManager) pollInterval() time.Duration {
	m.mu.Lock()
	count := len(m.liveGames)
	m.mu.Unlock()
	return max(3*time.Second, time.Duration(float64(count)/liveRequestBudget()*float64(time.Second)))
}

// ReplaceGames stops polling servers absent from the latest successful listing.
func (m *LiveGamesManager) ReplaceGames(games []model.Game) {
	present := map[string]bool{}
	for _, game := range games {
		present[game.ServerSteamID] = true
	}
	m.mu.Lock()
	for id, cancel := range m.liveGames {
		if !present[id] {
			cancel()
		}
	}
	m.mu.Unlock()
	for _, game := range games {
		m.AddGame(game)
	}
}
