// Package handler coordinates repositories and API loaders for HTTP requests.
package handler

import "time"

const cacheRefreshInterval = time.Hour

func cacheFresh(timestamp int64) bool {
	age := time.Since(time.Unix(timestamp, 0))
	return timestamp > 0 && age >= 0 && age < cacheRefreshInterval
}
