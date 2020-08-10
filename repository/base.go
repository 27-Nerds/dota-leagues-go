package repository

import (
	"context"
	"dota_league/db"
	e "dota_league/error"
	"time"
)

func existsInColByID(conn *db.Interface, colName string, id string) (bool, error) {
	query := "RETURN LENGTH(FOR d IN @@collection FILTER d._key == @id LIMIT 1 RETURN true) > 0"
	bindVars := map[string]interface{}{
		"@collection": colName,
		"id":          id,
	}

	var exists bool

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := (*conn).Query(ctx, query, bindVars, &exists)
	if e.IsNotFound(err) {
		// table not found. no need to crash
		return false, nil

	} else if err != nil {
		return false, &e.Error{Op: "repository.existsInColByID", Err: err}
	}

	return exists, nil
}
