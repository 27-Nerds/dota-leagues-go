package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"testing"
	"time"
)

type mockStandingsRepo struct {
	cached   *model.DPCStandings
	getErr   error
	stored   *model.DPCStandings
	storeErr error
}

func (m *mockStandingsRepo) Store(ctx context.Context, s *model.DPCStandings) error {
	m.stored = s
	return m.storeErr
}
func (m *mockStandingsRepo) Get(ctx context.Context) (*model.DPCStandings, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.cached, nil
}

// TestDPCHandlerReturnsCacheWithoutApiCall verifies a stored snapshot short-circuits the Valve call
func TestDPCHandlerReturnsCacheWithoutApiCall(t *testing.T) {
	repo := &mockStandingsRepo{cached: &model.DPCStandings{UpdatedTimestamp: time.Now().Unix()}}

	loaded := false
	h := NewDPCHandler(repo, func(ctx context.Context) (*model.DPCStandings, error) { loaded = true; return nil, nil })

	got, err := h.Get(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != repo.cached {
		t.Fatal("expected cached document to be returned")
	}
	if loaded {
		t.Fatal("api load must not run on cache hit")
	}
}

// TestDPCHandlerMissStoresAndReturnsApi verifies the fetch-then-store path
func TestDPCHandlerMissStoresAndReturnsApi(t *testing.T) {
	repo := &mockStandingsRepo{getErr: &e.Error{Code: e.ENOTFOUND, Message: "nf"}}

	apiData := &model.DPCStandings{}
	h := NewDPCHandler(repo, func(ctx context.Context) (*model.DPCStandings, error) { return apiData, nil })

	got, err := h.Get(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != apiData || repo.stored != apiData {
		t.Fatal("expected api data to be returned and stored")
	}
}

// TestDPCHandlerApiErrorPropagatesWithoutStore verifies a Valve failure is reported without a store attempt
func TestDPCHandlerApiErrorPropagatesWithoutStore(t *testing.T) {
	repo := &mockStandingsRepo{getErr: &e.Error{Code: e.ENOTFOUND, Message: "nf"}}

	apiErr := &e.Error{Code: e.EINTERNAL, Op: "api.LoadDPCStandings", Message: "502"}
	h := NewDPCHandler(repo, func(ctx context.Context) (*model.DPCStandings, error) { return nil, apiErr })

	got, err := h.Get(t.Context())
	if got != nil || err == nil {
		t.Fatalf("expected error and nil data, got %v / %v", got, err)
	}
	if repo.stored != nil {
		t.Fatal("nothing must be stored when the api call fails")
	}
}

// TestDPCHandlerStoreErrorStillServesApiData verifies db outages don't mask fresh data
func TestDPCHandlerStoreErrorStillServesApiData(t *testing.T) {
	repo := &mockStandingsRepo{getErr: &e.Error{Code: e.ENOTFOUND, Message: "nf"}, storeErr: &e.Error{Message: "db down"}}

	apiData := &model.DPCStandings{}
	h := NewDPCHandler(repo, func(ctx context.Context) (*model.DPCStandings, error) { return apiData, nil })

	got, err := h.Get(t.Context())
	if err != nil || got != apiData {
		t.Fatalf("expected fresh data despite store failure, got %v / %v", got, err)
	}
}

type mockResultsRepo struct {
	cached   *model.DPCLeagueResults
	getErr   error
	stored   *model.DPCLeagueResults
	storeErr error
}

func (m *mockResultsRepo) Store(ctx context.Context, leagueID int, res *model.DPCLeagueResults) error {
	m.stored = res
	return m.storeErr
}
func (m *mockResultsRepo) Get(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.cached, nil
}

// TestDPCResultsHandlerCacheHit verifies cached league results skip the api call
func TestDPCResultsHandlerCacheHit(t *testing.T) {
	repo := &mockResultsRepo{cached: &model.DPCLeagueResults{UpdatedTimestamp: time.Now().Unix()}}

	loaded := false
	h := NewDPCResultsHandler(repo, func(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error) {
		loaded = true
		return nil, nil
	})

	got, err := h.Get(t.Context(), 123)
	if err != nil || got != repo.cached {
		t.Fatalf("expected cached data, got %v / %v", got, err)
	}
	if loaded {
		t.Fatal("api load must not run on cache hit")
	}
}

// TestDPCResultsHandlerMissStores verifies the fetch-then-store path for league results
func TestDPCResultsHandlerMissStores(t *testing.T) {
	repo := &mockResultsRepo{getErr: &e.Error{Code: e.ENOTFOUND, Message: "nf"}}

	apiData := &model.DPCLeagueResults{}
	h := NewDPCResultsHandler(repo, func(ctx context.Context, leagueID int) (*model.DPCLeagueResults, error) { return apiData, nil })

	got, err := h.Get(t.Context(), 4075214)
	if err != nil || got != apiData || repo.stored != apiData {
		t.Fatalf("expected fetch+store of %v, got stored=%p err=%v", got, repo.stored, err)
	}
}

type mockMatchRepo struct {
	cached   *model.MatchMinimal
	getErr   error
	stored   *model.MatchMinimal
	storeErr error
}

func (m *mockMatchRepo) Store(ctx context.Context, mm *model.MatchMinimal) error {
	m.stored = mm
	return m.storeErr
}
func (m *mockMatchRepo) Get(ctx context.Context, matchID string) (*model.MatchMinimal, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.cached, nil
}

// TestMatchMinimalHandlerCacheHit verifies a stored match short-circuits the api call
func TestMatchMinimalHandlerCacheHit(t *testing.T) {
	repo := &mockMatchRepo{cached: &model.MatchMinimal{}}

	loaded := false
	h := NewMatchMinimalHandler(repo, func(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
		loaded = true
		return nil, nil
	})

	got, err := h.Get(t.Context(), 1, "7105932934")
	if err != nil || got != repo.cached {
		t.Fatalf("expected cached data, got %v / %v", got, err)
	}
	if loaded {
		t.Fatal("api load must not run on cache hit")
	}
}

// TestMatchMinimalHandlerMissStores verifies the fetch-then-store path for matches
func TestMatchMinimalHandlerMissStores(t *testing.T) {
	repo := &mockMatchRepo{getErr: &e.Error{Code: e.ENOTFOUND, Message: "nf"}}

	apiData := &model.MatchMinimal{}
	h := NewMatchMinimalHandler(repo, func(ctx context.Context, leagueID int, matchID string) (*model.MatchMinimal, error) {
		return apiData, nil
	})

	got, err := h.Get(t.Context(), 1, "7105932934")
	if err != nil || got != apiData || repo.stored != apiData {
		t.Fatalf("expected fetch+store of %v, got stored=%p err=%v", got, repo.stored, err)
	}
}

func TestDPCHandlerRefreshesExpiredCache(t *testing.T) {
	for _, timestamp := range []int64{0, time.Now().Add(-2 * time.Hour).Unix()} {
		repo := &mockStandingsRepo{cached: &model.DPCStandings{UpdatedTimestamp: timestamp}}
		fresh := &model.DPCStandings{}
		h := NewDPCHandler(repo, func(ctx context.Context) (*model.DPCStandings, error) { return fresh, nil })
		got, err := h.Get(t.Context())
		if err != nil || got != fresh || repo.stored != fresh || !cacheFresh(fresh.UpdatedTimestamp) {
			t.Fatalf("timestamp %d: expected refreshed and stored snapshot, got %v / %v", timestamp, got, err)
		}
	}
}

func TestDPCResultsHandlerRefreshesExpiredCache(t *testing.T) {
	for _, timestamp := range []int64{0, time.Now().Add(-2 * time.Hour).Unix()} {
		repo := &mockResultsRepo{cached: &model.DPCLeagueResults{UpdatedTimestamp: timestamp}}
		fresh := &model.DPCLeagueResults{}
		h := NewDPCResultsHandler(repo, func(context.Context, int) (*model.DPCLeagueResults, error) { return fresh, nil })
		got, err := h.Get(t.Context(), 123)
		if err != nil || got != fresh || repo.stored != fresh || !cacheFresh(fresh.UpdatedTimestamp) {
			t.Fatalf("timestamp %d: expected refreshed and stored snapshot, got %v / %v", timestamp, got, err)
		}
	}
}
