package auth

import (
	"sync"
	"time"
)

// TTLCache provides a generic thread-safe cache with time-to-live expiration.
// It can be used to cache any type of data with automatic expiration.
type TTLCache[K comparable, V any] struct {
	items      map[K]cacheItem[V]
	mutex      sync.RWMutex
	defaultTTL time.Duration
}

// cacheItem wraps a cached value with its expiration time.
type cacheItem[V any] struct {
	Value     V
	ExpiresAt time.Time
}

// NewTTLCache creates a new TTL cache with the specified default TTL.
func NewTTLCache[K comparable, V any](defaultTTL time.Duration) *TTLCache[K, V] {
	return &TTLCache[K, V]{
		items:      make(map[K]cacheItem[V]),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves a value from cache if it exists and hasn't expired.
// Returns (value, true) if found and valid, (zero, false) otherwise.
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		var zero V
		return zero, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		var zero V
		return zero, false
	}

	return item.Value, true
}

// Set stores a value in cache with the default TTL.
func (c *TTLCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in cache with a custom TTL.
func (c *TTLCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = cacheItem[V]{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a specific key from cache.
func (c *TTLCache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from cache.
func (c *TTLCache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[K]cacheItem[V])
}

// ClearExpired removes all expired items from cache.
// Should be called periodically to prevent memory leaks.
func (c *TTLCache[K, V]) ClearExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}


// Size returns the current number of items in cache (including expired ones).
func (c *TTLCache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// StringTTLCache is a specialized TTL cache for string keys.
// It provides additional methods like DeleteByPrefix that are specific to string keys.
type StringTTLCache[V any] struct {
	*TTLCache[string, V]
}

// NewStringTTLCache creates a new TTL cache with string keys.
func NewStringTTLCache[V any](defaultTTL time.Duration) *StringTTLCache[V] {
	return &StringTTLCache[V]{
		TTLCache: NewTTLCache[string, V](defaultTTL),
	}
}

// DeleteByPrefix removes all keys that start with the given prefix.
// This is useful for invalidating related cache entries (e.g., all entries for a user).
func (c *StringTTLCache[V]) DeleteByPrefix(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key := range c.items {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(c.items, key)
		}
	}
}
