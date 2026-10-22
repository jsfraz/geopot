package utils

import (
	"errors"
	"testing"
	"time"
)

func TestTTLCacheLoadsOnceWithinTTL(t *testing.T) {
	now := time.Unix(0, 0)
	cache := NewTTLCache[int]()
	cache.now = func() time.Time { return now }

	calls := 0
	load := func() (int, error) {
		calls++
		return 42, nil
	}

	for i := 0; i < 3; i++ {
		value, err := cache.Get("key", time.Minute, 0, load)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != 42 {
			t.Errorf("value = %d, want 42", value)
		}
	}
	if calls != 1 {
		t.Errorf("load called %d times, want 1", calls)
	}

	// After the TTL the value is loaded again.
	now = now.Add(time.Minute + time.Second)
	if _, err := cache.Get("key", time.Minute, 0, load); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 2 {
		t.Errorf("load called %d times, want 2", calls)
	}
}

func TestTTLCacheServesStaleValueOnError(t *testing.T) {
	now := time.Unix(0, 0)
	cache := NewTTLCache[int]()
	cache.now = func() time.Time { return now }

	if _, err := cache.Get("key", time.Minute, 5*time.Minute, func() (int, error) { return 7, nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expired but within the stale window: the previous value is returned.
	now = now.Add(2 * time.Minute)
	value, err := cache.Get("key", time.Minute, 5*time.Minute, func() (int, error) {
		return 0, errors.New("database is down")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 7 {
		t.Errorf("value = %d, want the stale 7", value)
	}

	// Beyond the stale window the error is propagated.
	now = now.Add(10 * time.Minute)
	if _, err := cache.Get("key", time.Minute, 5*time.Minute, func() (int, error) {
		return 0, errors.New("database is down")
	}); err == nil {
		t.Error("expected an error once the stale window is over")
	}
}

func TestTTLCacheKeysAreIndependent(t *testing.T) {
	cache := NewTTLCache[string]()

	first, err := cache.Get("a", time.Minute, 0, func() (string, error) { return "A", nil })
	if err != nil || first != "A" {
		t.Fatalf("first = %q, err = %v", first, err)
	}
	second, err := cache.Get("b", time.Minute, 0, func() (string, error) { return "B", nil })
	if err != nil || second != "B" {
		t.Fatalf("second = %q, err = %v", second, err)
	}
}
