package db

import (
	"context"

	arango "github.com/arangodb/go-driver"
)

// Interface database interface
type Interface interface {
	Insert(ctx context.Context, colName string, obj interface{}) error
	InsertMany(ctx context.Context, colName string, obj interface{}) error
	Query(ctx context.Context, query string, bindVars map[string]interface{}, resObj interface{}) (string, error)
	// this query breaks all the flexibility. lets wait for generics
	QueryAll(ctx context.Context, query string, bindVars map[string]interface{}) (arango.Cursor, error)

	Update(ctx context.Context, colName string, key string, obj interface{}) error
	DoQuery(ctx context.Context, query string) error
	ClearCollection(ctx context.Context, colName string) error
}
