package utils

import (
	"sync"
	"time"
)

// RateLimiter is a small per-key fixed-window rate limiter.
type RateLimiter struct {
	mutex  sync.Mutex
	limit  int
	window time.Duration
	now    func() time.Time
	items  map[string]*rateLimitEntry
}

type rateLimitEntry struct {
	count int
	reset time.Time
}

// Creates a new rate limiter allowing limit requests per minute per key.
//
//	@param limit
//	@return *RateLimiter
func NewRateLimiter(limit int) *RateLimiter {
	if limit <= 0 {
		limit = 120
	}
	return &RateLimiter{
		limit:  limit,
		window: time.Minute,
		now:    time.Now,
		items:  make(map[string]*rateLimitEntry),
	}
}

// Reports whether another request for key is allowed within the current window.
//
//	@receiver r
//	@param key
//	@return bool
func (r *RateLimiter) Allow(key string) bool {
	now := r.now()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Opportunistically drop expired entries to keep memory bounded.
	if len(r.items) > 10000 {
		for k, entry := range r.items {
			if now.After(entry.reset) {
				delete(r.items, k)
			}
		}
	}

	entry, ok := r.items[key]
	if !ok || now.After(entry.reset) {
		r.items[key] = &rateLimitEntry{count: 1, reset: now.Add(r.window)}
		return true
	}

	if entry.count >= r.limit {
		return false
	}
	entry.count++
	return true
}
