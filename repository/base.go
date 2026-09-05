// Package repository stores and retrieves application data in ArangoDB.
package repository

import (
	"context"
	e "dota_league/error"
)

func existsInColByID(ctx context.Context, conn Database, colName string, id string) (bool, error) {
	query := "RETURN LENGTH(FOR d IN @@collection FILTER d._key == @id LIMIT 1 RETURN true) > 0"
	bindVars := map[string]any{
		"@collection": colName,
		"id":          id,
	}

	var exists bool

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()
	_, err := conn.Query(ctx, query, bindVars, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
		return false, &e.Error{Op: "repository.existsInColByID", Err: err}
	}

	return exists, nil
}
