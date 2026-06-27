package cache

import (
	"sync"
	"time"
)

// Cache is a simple in-memory cache with TTL.
type Cache struct {
	mu    sync.RWMutex
	items map[string]Item
}

// Item represents a cached item.
type Item struct {
	Value     any
	ExpiresAt time.Time
}

// New creates a new cache.
func New() *Cache {
	return &Cache{
		items: make(map[string]Item),
	}
}

// Get gets a value from the cache.
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Value, true
}

// Set sets a value in the cache with TTL.
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete deletes a value from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear clears all items from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]Item)
}

// GetOrSet gets a value from the cache, or sets it if not present.
func (c *Cache) GetOrSet(key string, ttl time.Duration, fn func() (any, error)) (any, error) {
	// Try to get
	if val, ok := c.Get(key); ok {
		return val, nil
	}

	// Compute value
	val, err := fn()
	if err != nil {
		return nil, err
	}

	// Set in cache
	c.Set(key, val, ttl)

	return val, nil
}

// Size returns the number of items in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// Cleanup removes expired items.
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.ExpiresAt) {
			delete(c.items, key)
		}
	}
}

// StartCleanup starts a background goroutine to cleanup expired items.
func (c *Cache) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			c.Cleanup()
		}
	}()
}
