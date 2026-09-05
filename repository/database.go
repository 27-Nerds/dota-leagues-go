package repository

import (
	"context"

	arango "github.com/arangodb/go-driver/v2/arangodb"
)

// Database provides the ArangoDB operations used by repositories.
type Database interface {
	WithTransaction(ctx context.Context, colName string, fn func(context.Context) error) error
	Insert(ctx context.Context, colName string, obj any) error
	InsertMany(ctx context.Context, colName string, obj any) error
	Query(ctx context.Context, query string, bindVars map[string]any, resObj any) (string, error)
	QueryAll(ctx context.Context, query string, bindVars map[string]any, fullCount bool) (arango.Cursor, error)

	Update(ctx context.Context, colName string, key string, obj any) error
	DoQuery(ctx context.Context, query string) error
	DoQueryBuilder(ctx context.Context, query string, bindVars map[string]any) error
	ClearCollection(ctx context.Context, colName string) error
}
