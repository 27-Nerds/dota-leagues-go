package db

import (
	"context"
	e "dota_league/error"
	"fmt"

	arango "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
)

// ArangoDb struct
type ArangoDb struct {
	Db *arango.Database
}

// Connect to database
func Connect(ctx context.Context,
	dbURL string,
	dbUser string,
	dbPass string,
	dbName string,
) (Interface, error) {
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
	} else if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return &ArangoDb{&db}, nil
}

// Collection create and get collection
func (a *ArangoDb) collection(ctx context.Context, colName string) (arango.Collection, error) {
	var col arango.Collection

	col, err := (*a.Db).Collection(ctx, colName)

	if arango.IsNotFound(err) {
		col, err = (*a.Db).CreateCollection(ctx, colName, nil)
	} else if err != nil {
		return nil, &e.Error{Code: e.EINTERNAL, Op: "db.collection", Err: err}
	}

	return col, nil
}

// Query assing result to resObj and returns id as first value
func (a *ArangoDb) Query(ctx context.Context, query string, bindVars map[string]interface{}, resObj interface{}) (string, error) {
	const op = "db.Query"
	cursor, err := (*a.Db).Query(ctx, query, bindVars)

	//collection not found
	if arango.IsNotFound(err) {
		return "", &e.Error{Code: e.ENOTFOUND, Op: op}
	} else if err != nil {
		// handle error
		return "", &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}
	defer cursor.Close()

	meta, err := cursor.ReadDocument(ctx, &resObj)
	if arango.IsNoMoreDocuments(err) {
		return "", &e.Error{Code: e.ENOTFOUND, Op: op}
	} else if err != nil {
		return "", &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return meta.Key, nil
}

// Update document
func (a *ArangoDb) Update(ctx context.Context, colName string, key string, obj interface{}) error {
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
func (a *ArangoDb) Insert(ctx context.Context, colName string, obj interface{}) error {
	const op = "db.Insert"
	col, err := a.collection(ctx, colName)
	if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	_, err = col.CreateDocument(ctx, obj)
	if arango.IsPreconditionFailed(err) {
		//Index already exist in the DB
		return &e.Error{Code: e.ECONFLICT, Op: op, Err: err}
	} else if err != nil {
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}

// InsertMany - batch insert values
func (a *ArangoDb) InsertMany(ctx context.Context, colName string, obj interface{}) error {
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
func (a *ArangoDb) QueryAll(ctx context.Context, query string, bindVars map[string]interface{}) (arango.Cursor, error) {
	const op = "db.QueryAll"

	cursor, err := (*a.Db).Query(ctx, query, bindVars)
	if err != nil {
		// handle error
		return nil, &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return cursor, nil
}

// ClearCollection clear data from the collection
func (a *ArangoDb) ClearCollection(ctx context.Context, colName string) error {
	const op = "db.ClearCollection"

	query := fmt.Sprintf("FOR u IN %s REMOVE u IN %s", colName, colName)

	_, err := (*a.Db).Query(ctx, query, nil)
	if err != nil {
		// handle error
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}

// DoQuery can be used to perform any arango query
func (a *ArangoDb) DoQuery(ctx context.Context, query string) error {

	const op = "db.DoUpdateQuery"

	_, err := (*a.Db).Query(ctx, query, nil)
	if err != nil {
		// handle error
		return &e.Error{Code: e.EINTERNAL, Op: op, Err: err}
	}

	return nil
}
