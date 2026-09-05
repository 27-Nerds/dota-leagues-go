// Package db manages ArangoDB connections, queries, and transactions.
package db

import (
	"context"
	e "dota_league/error"
	"fmt"
	"log/slog"
	"time"

	arango "github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
	"github.com/arangodb/go-driver/v2/connection"
)

// ArangoDB struct
type ArangoDB struct {
	DB arango.Database
}

// Connect to database
func Connect(ctx context.Context,
	dbURL string,
	dbUser string,
	dbPass string,
	dbName string,
) (*ArangoDB, error) {
	const op = "db.Connect"

	conn := connection.NewHttpConnection(connection.HttpConfiguration{
		Endpoint:       connection.NewRoundRobinEndpoints([]string{dbURL}),
		Authentication: connection.NewBasicAuth(dbUser, dbPass),
	})
	c := arango.NewClient(conn)

	db, err := c.GetDatabase(ctx, dbName, nil)
	if shared.IsNotFound(err) {
		db, err = c.CreateDatabase(ctx, dbName, nil)
	}
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &ArangoDB{DB: db}, nil
}

// Collection create and get collection
func (a *ArangoDB) collection(ctx context.Context, colName string) (arango.Collection, error) {
	var col arango.Collection

	if tx, ok := ctx.Value(transactionKey{}).(arango.Transaction); ok {
		return tx.GetCollection(ctx, colName, &arango.GetCollectionOptions{SkipExistCheck: true})
	}
	col, err := a.DB.GetCollection(ctx, colName, nil)

	if shared.IsNotFound(err) {
		col, err = a.DB.CreateCollectionV2(ctx, colName, nil)
	}
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: "db.collection", Err: err}
	}

	return col, nil
}

// Query assing result to resObj and returns id as first value
func (a *ArangoDB) Query(ctx context.Context, query string, bindVars map[string]any, resObj any) (string, error) {
	const op = "db.Query"
	cursor, err := a.query(ctx, query, bindVars, false)

	//collection not found
	if shared.IsNotFound(err) {
		return "", &e.Error{Code: e.ENOTFOUND, Op: op}
	} else if err != nil {
		// handle error
		return "", &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer CloseCursor(cursor)

	meta, err := cursor.ReadDocument(ctx, resObj)
	if shared.IsNoMoreDocuments(err) {
		return "", &e.Error{Code: e.ENOTFOUND, Op: op}
	} else if err != nil {
		return "", &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return meta.Key, nil
}

// Update document
func (a *ArangoDB) Update(ctx context.Context, colName string, key string, obj any) error {
	const op = "db.Update"
	col, err := a.collection(ctx, colName)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	_, err = col.UpdateDocument(ctx, key, obj)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}

// Insert value
func (a *ArangoDB) Insert(ctx context.Context, colName string, obj any) error {
	const op = "db.Insert"
	col, err := a.collection(ctx, colName)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	_, err = col.CreateDocument(ctx, obj)
	if shared.IsPreconditionFailed(err) || shared.IsConflict(err) {
		//document with the same _key already exist in the DB (412 or 409/1210)
		return &e.Error{Code: e.ECONFLICT, Op: op, Err: err}
	} else if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}

// InsertMany - batch insert values
func (a *ArangoDB) InsertMany(ctx context.Context, colName string, obj any) error {
	const op = "db.InsertMany"
	col, err := a.collection(ctx, colName)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	reader, err := col.CreateDocuments(ctx, obj)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	// Batch requests can succeed while individual documents fail.
	for {
		_, err := reader.Read()
		if shared.IsNoMoreDocuments(err) {
			break
		}
		if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}
	}
	return nil
}

// QueryAll returns a cursor, optionally including the total before pagination.
func (a *ArangoDB) QueryAll(ctx context.Context, query string, bindVars map[string]any, fullCount bool) (arango.Cursor, error) {
	const op = "db.QueryAll"

	cursor, err := a.query(ctx, query, bindVars, fullCount)
	if err != nil {
		// handle error
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return cursor, nil
}

// ClearCollection clear data from the collection
func (a *ArangoDB) ClearCollection(ctx context.Context, colName string) error {
	query := fmt.Sprintf("FOR u IN %s REMOVE u IN %s", colName, colName)
	return a.DoQuery(ctx, query)
}

// DoQuery can be used to perform any arango query
func (a *ArangoDB) DoQuery(ctx context.Context, query string) error {
	return a.DoQueryBuilder(ctx, query, nil)
}

// DoQueryBuilder runs an arbitrary arango query with bind variables.
// The cursor is drained so that mutating queries (INSERT/REMOVE/REPLACE) actually execute.
func (a *ArangoDB) DoQueryBuilder(ctx context.Context, query string, bindVars map[string]any) error {
	const op = "db.DoQueryBuilder"

	cursor, err := a.query(ctx, query, bindVars, false)
	if err != nil {
		// handle error
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer CloseCursor(cursor)

	for {
		var doc any
		_, err := cursor.ReadDocument(ctx, &doc)
		if shared.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
		}
	}

	return nil
}

// WithTransaction runs fn atomically against one collection, creating it first if needed.
// All operations in fn must use the supplied transaction context.
func (a *ArangoDB) WithTransaction(ctx context.Context, colName string, fn func(context.Context) error) error {
	const op = "db.WithTransaction"
	if _, err := a.collection(ctx, colName); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	tx, err := a.DB.BeginTransaction(ctx, arango.TransactionCollections{Write: []string{colName}}, nil)
	if err != nil {
		return &e.Error{Op: op, Err: err}
	}
	committed := false
	defer func() {
		if !committed {
			// The request context may have expired; rollback still needs a deadline.
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tx.Abort(cleanupCtx, nil); err != nil {
				slog.Error("abort transaction", "transaction_id", tx.ID(), "error", err)
			}
		}
	}()
	if err := fn(context.WithValue(ctx, transactionKey{}, tx)); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	if err := tx.Commit(ctx, nil); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	committed = true
	return nil
}

// CloseCursor releases a query cursor and reports cleanup failures.
func CloseCursor(cursor arango.Cursor) {
	if err := cursor.Close(); err != nil {
		slog.Warn("close database cursor", "error", err)
	}
}

// transactionKey keeps transaction state local to the operation, not the shared adapter.
type transactionKey struct{}

func (a *ArangoDB) query(ctx context.Context, query string, bindVars map[string]any, fullCount bool) (arango.Cursor, error) {
	var executor arango.DatabaseQuery = a.DB
	if tx, ok := ctx.Value(transactionKey{}).(arango.Transaction); ok {
		executor = tx
	}
	return executor.Query(ctx, query, &arango.QueryOptions{
		BindVars: bindVars,
		Options:  arango.QuerySubOptions{FullCount: fullCount},
	})
}
