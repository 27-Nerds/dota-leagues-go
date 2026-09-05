package worker

import (
	"context"
	"dota_league/api"
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
	liveGames map[string]struct{}
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
		liveGames: make(map[string]struct{}),
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
	m.liveGames[id] = struct{}{}
	// Register the task while holding the same lock Stop uses before waiting.
	m.wg.Go(func() { m.poll(id) })
}

// Stop prevents new games, cancels in-flight requests, and joins all polling workers.
func (m *LiveGamesManager) Stop() {
	m.mu.Lock()
	m.stopped = true
	m.cancel()
	m.mu.Unlock()
	m.wg.Wait()
}

func (m *LiveGamesManager) poll(id string) {
	defer func() {
		m.mu.Lock()
		delete(m.liveGames, id)
		m.mu.Unlock()
	}()
	deadline := time.Now().Add(liveGameTimeout)
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-timer.C:
		}
		if m.ctx.Err() != nil || !time.Now().Before(deadline) {
			return
		}
		// The inactivity deadline also bounds a stalled request or database write.
		ctx, cancel := context.WithDeadline(m.ctx, deadline)
		details, err := m.load(ctx, id)
		if err == nil {
			err = m.store(ctx, details)
		}
		cancel()
		if err == nil {
			deadline = time.Now().Add(liveGameTimeout)
		} else if m.ctx.Err() == nil {
			slog.WarnContext(m.ctx, "live game refresh failed", "server_id", id, "error", err)
		}
		timer.Reset(min(3*time.Second, max(0, time.Until(deadline))))
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
