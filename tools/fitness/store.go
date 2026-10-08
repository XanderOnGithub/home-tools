package fitness

import "sync"

type Store struct {
	dir       string
	mu        sync.RWMutex
	exercises map[string]Exercise // by ID
	routines  map[string]Routine  // by ID
	sessions  map[string]Session  // by userID, sorted by date (oldest -> newest)
}
