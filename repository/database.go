package repository

import (
	"context"

	arango "github.com/arangodb/go-driver"
)

// Database provides the ArangoDB operations used by repositories.
type Database interface {
	WithTransaction(ctx context.Context, colName string, fn func(context.Context) error) error
	Insert(ctx context.Context, colName string, obj interface{}) error
	InsertMany(ctx context.Context, colName string, obj interface{}) error
	Query(ctx context.Context, query string, bindVars map[string]interface{}, resObj interface{}) (string, error)
	QueryAll(ctx context.Context, query string, bindVars map[string]interface{}) (arango.Cursor, error)

	Update(ctx context.Context, colName string, key string, obj interface{}) error
	DoQuery(ctx context.Context, query string) error
	DoQueryBuilder(ctx context.Context, query string, bindVars map[string]interface{}) error
	ClearCollection(ctx context.Context, colName string) error
}
