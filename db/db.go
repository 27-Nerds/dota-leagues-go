// Package db manages ArangoDB connections, queries, and transactions.
package db

import (
	"context"
	e "dota_league/error"
	"fmt"
	"log"
	"time"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
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

	var db arango.Database
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{dbURL},
	})
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	c, err := arango.NewClient(arango.ClientConfig{
		Connection:     conn,
		Authentication: arango.BasicAuthentication(dbUser, dbPass),
	})
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	db, err = c.Database(ctx, dbName)
	if arango.IsNotFound(err) {
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

	col, err := a.DB.Collection(ctx, colName)

	if arango.IsNotFound(err) {
		col, err = a.DB.CreateCollection(ctx, colName, nil)
	}
	if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: "db.collection", Err: err}
	}

	return col, nil
}

// Query assing result to resObj and returns id as first value
func (a *ArangoDB) Query(ctx context.Context, query string, bindVars map[string]interface{}, resObj interface{}) (string, error) {
	const op = "db.Query"
	cursor, err := a.DB.Query(ctx, query, bindVars)

	//collection not found
	if arango.IsNotFound(err) {
		return "", &e.Error{Code: e.ENOTFOUND, Op: op}
	} else if err != nil {
		// handle error
		return "", &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer CloseCursor(cursor)

	meta, err := cursor.ReadDocument(ctx, &resObj)
	if arango.IsNoMoreDocuments(err) {
		return "", &e.Error{Code: e.ENOTFOUND, Op: op}
	} else if err != nil {
		return "", &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return meta.Key, nil
}

// Update document
func (a *ArangoDB) Update(ctx context.Context, colName string, key string, obj interface{}) error {
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
func (a *ArangoDB) Insert(ctx context.Context, colName string, obj interface{}) error {
	const op = "db.Insert"
	col, err := a.collection(ctx, colName)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	_, err = col.CreateDocument(ctx, obj)
	if arango.IsPreconditionFailed(err) || arango.IsConflict(err) {
		//document with the same _key already exist in the DB (412 or 409/1210)
		return &e.Error{Code: e.ECONFLICT, Op: op, Err: err}
	} else if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}

// InsertMany - batch insert values
func (a *ArangoDB) InsertMany(ctx context.Context, colName string, obj interface{}) error {
	const op = "db.InsertMany"
	col, err := a.collection(ctx, colName)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	_, errs, err := col.CreateDocuments(ctx, obj)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	} else if err := errs.FirstNonNil(); err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}

// QueryAll return cursor as we do not support generics
func (a *ArangoDB) QueryAll(ctx context.Context, query string, bindVars map[string]interface{}) (arango.Cursor, error) {
	const op = "db.QueryAll"

	cursor, err := a.DB.Query(ctx, query, bindVars)
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
func (a *ArangoDB) DoQueryBuilder(ctx context.Context, query string, bindVars map[string]interface{}) error {
	const op = "db.DoQueryBuilder"

	cursor, err := a.DB.Query(ctx, query, bindVars)
	if err != nil {
		// handle error
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer CloseCursor(cursor)

	for {
		var doc interface{}
		_, err := cursor.ReadDocument(ctx, &doc)
		if arango.IsNoMoreDocuments(err) {
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
	id, err := a.DB.BeginTransaction(ctx, arango.TransactionCollections{Write: []string{colName}}, nil)
	if err != nil {
		return &e.Error{Op: op, Err: err}
	}
	committed := false
	defer func() {
		if !committed {
			// The request context may have expired; rollback still needs a deadline.
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := a.DB.AbortTransaction(cleanupCtx, id, nil); err != nil {
				log.Printf("abort transaction %s: %v", id, err)
			}
		}
	}()
	if err := fn(arango.WithTransactionID(ctx, id)); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	if err := a.DB.CommitTransaction(ctx, id, nil); err != nil {
		return &e.Error{Op: op, Err: err}
	}
	committed = true
	return nil
}

// CloseCursor releases a query cursor and reports cleanup failures.
func CloseCursor(cursor arango.Cursor) {
	if err := cursor.Close(); err != nil {
		log.Printf("close database cursor: %v", err)
	}
}
