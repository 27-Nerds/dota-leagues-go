package repository

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
	"testing"

	arango "github.com/arangodb/go-driver"
)

// fakeDB records calls and lets tests script Insert/Update outcomes
type fakeDB struct {
	insertCalls int
	updateCalls int
	lastInsert  interface{}
	lastUpdate  interface{}
	lastKey     string
	insertErr   error
	updateErr   error
}

func (f *fakeDB) Insert(ctx context.Context, colName string, obj interface{}) error {
	f.insertCalls++
	f.lastInsert = obj
	return f.insertErr
}

func (f *fakeDB) InsertMany(ctx context.Context, colName string, obj interface{}) error {
	f.insertCalls += 1
	f.lastInsert = obj
	return nil
}

func (f *fakeDB) Query(ctx context.Context, query string, bindVars map[string]interface{}, resObj interface{}) (string, error) {
	return "", &e.Error{Code: e.ENOTFOUND, Message: "no document found"}
}

func (f *fakeDB) QueryAll(ctx context.Context, query string, bindVars map[string]interface{}) (arango.Cursor, error) {
	return nil, &e.Error{Code: e.ENOTFOUND, Message: "no documents found"}
}

func (f *fakeDB) Update(ctx context.Context, colName string, key string, obj interface{}) error {
	f.updateCalls++
	f.lastKey = key
	f.lastUpdate = obj
	return f.updateErr
}

func (f *fakeDB) DoQuery(ctx context.Context, query string) error { return nil }
func (f *fakeDB) DoQueryBuilder(ctx context.Context, query string, bindVars map[string]interface{}) error {
	return nil
}
func (f *fakeDB) ClearCollection(ctx context.Context, colName string) error { return nil }

var econflictErr = &e.Error{Code: e.ECONFLICT, Message: "conflict"}

// TestDPCStandingsStoreInsertHappyPath stores via insert on a fresh key
func TestDPCStandingsStoreInsertHappyPath(t *testing.T) {
	fake := &fakeDB{}
	repo := NewDPCStandingsRepository(fake)

	err := repo.Store(&model.DPCStandings{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fake.insertCalls != 1 || fake.updateCalls != 0 {
		t.Fatalf("expected insert=1 update=0, got insert=%d update=%d", fake.insertCalls, fake.updateCalls)
	}
	std := fake.lastInsert.(*model.DPCStandings)
	if std.DBKey != dpcKey {
		t.Fatalf("expected db key %q, got %q", dpcKey, std.DBKey)
	}
}

// TestDPCStandingsStoreConflictUpdates verifies a duplicate _key falls back to update instead of erroring
func TestDPCStandingsStoreConflictUpdates(t *testing.T) {
	fake := &fakeDB{insertErr: econflictErr}
	repo := NewDPCStandingsRepository(fake)

	err := repo.Store(&model.DPCStandings{})
	if err != nil {
		t.Fatalf("expected conflict to resolve via update, got %v", err)
	}
	if fake.updateCalls != 1 || fake.lastKey != dpcKey {
		t.Fatalf("expected single update for key %q, got calls=%d key=%q", dpcKey, fake.updateCalls, fake.lastKey)
	}
}

// TestDPCStandingsStoreOtherErrorPropagates ensures non-conflict errors are not swallowed
func TestDPCStandingsStoreOtherErrorPropagates(t *testing.T) {
	fake := &fakeDB{insertErr: &e.Error{Code: e.EINTERNAL, Message: "boom"}}
	repo := NewDPCStandingsRepository(fake)

	if err := repo.Store(&model.DPCStandings{}); err == nil {
		t.Fatal("expected error to propagate")
	}
	if fake.updateCalls != 0 {
		t.Fatalf("no update expected, got %d", fake.updateCalls)
	}
}

// TestDPCResultsStoreConflictUpdates verifies the per-league upsert conflict path
func TestDPCResultsStoreConflictUpdates(t *testing.T) {
	fake := &fakeDB{insertErr: econflictErr}
	repo := NewDPCResultsRepository(fake)

	err := repo.Store(4075214, &model.DPCLeagueResults{})
	if err != nil {
		t.Fatalf("expected conflict to resolve via update, got %v", err)
	}
	if fake.updateCalls != 1 || fake.lastKey != "4075214" {
		t.Fatalf("expected single update for key 4075214, got calls=%d key=%q", fake.updateCalls, fake.lastKey)
	}
	res := fake.lastInsert.(*model.DPCLeagueResults)
	if res.DBKey != "4075214" {
		t.Fatalf("expected db key 4075214, got %q", res.DBKey)
	}
}

// TestMatchMinimalStoreConflictUpdates verifies the per-match upsert conflict path
func TestMatchMinimalStoreConflictUpdates(t *testing.T) {
	fake := &fakeDB{insertErr: econflictErr}
	repo := NewMatchMinimalRepository(fake)

	err := repo.Store(&model.MatchMinimal{MatchID: "7105932934"})
	if err != nil {
		t.Fatalf("expected conflict to resolve via update, got %v", err)
	}
	if fake.updateCalls != 1 || fake.lastKey != "7105932934" {
		t.Fatalf("expected single update for key 7105932934, got calls=%d key=%q", fake.updateCalls, fake.lastKey)
	}
	mm := fake.lastInsert.(*model.MatchMinimal)
	if mm.DBKey != "7105932934" {
		t.Fatalf("expected db key 7105932934, got %q", mm.DBKey)
	}
}

func (f *fakeDB) WithTransaction(ctx context.Context, colName string, fn func(context.Context) error) error {
	return fn(ctx)
}
