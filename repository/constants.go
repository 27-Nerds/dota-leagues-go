package repository

import "time"

// dbTimeout bounds every database operation in this package
const dbTimeout = 5 * time.Second
