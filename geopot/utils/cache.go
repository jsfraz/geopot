package utils

import (
	"sync"
	"time"
)

// TTLCache is a small in-process cache for expensive read-only queries. It
// caches one value per key, keeps it for ttl and, when a refresh fails, keeps
// serving the previous value for an additional staleFor window.
//
// It is meant for a bounded set of keys (one per query), not as a general
// purpose cache: entries are never evicted individually.
type TTLCache[T any] struct {
	mutex   sync.Mutex
	entries map[string]*cacheEntry[T]
	now     func() time.Time
}

type cacheEntry[T any] struct {
	mutex    sync.Mutex
	value    T
	loadedAt time.Time
	hasValue bool
}

// Creates an empty cache.
//
//	@return *TTLCache
func NewTTLCache[T any]() *TTLCache[T] {
	return &TTLCache[T]{
		entries: make(map[string]*cacheEntry[T]),
		now:     time.Now,
	}
}

// Returns the cached value for key, loading it with load when the entry is
// missing or older than ttl. Concurrent callers asking for the same key share a
// single load call.
//
//	@receiver c
//	@param key
//	@param ttl
//	@param staleFor
//	@param load
//	@return T
//	@return error
func (c *TTLCache[T]) Get(key string, ttl time.Duration, staleFor time.Duration, load func() (T, error)) (T, error) {
	c.mutex.Lock()
	entry, ok := c.entries[key]
	if !ok {
		entry = &cacheEntry[T]{}
		c.entries[key] = entry
	}
	c.mutex.Unlock()

	entry.mutex.Lock()
	defer entry.mutex.Unlock()

	now := c.now()
	if entry.hasValue && now.Sub(entry.loadedAt) < ttl {
		return entry.value, nil
	}

	value, err := load()
	if err != nil {
		// Prefer slightly outdated data over an error page while the previous
		// value is still within the stale window.
		if entry.hasValue && now.Sub(entry.loadedAt) < ttl+staleFor {
			return entry.value, nil
		}
		var zero T
		return zero, err
	}

	entry.value = value
	entry.loadedAt = now
	entry.hasValue = true
	return value, nil
}
