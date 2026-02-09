package main

import (
	"sync"
	"time"

	"dramawave/models"
)

const (
	CacheTTL = 6 * time.Hour
)

type CacheEntry struct {
	Data      *models.SeriesDetailResponse
	ExpiresAt time.Time
}

type SeriesCache struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry
}

func NewSeriesCache() *SeriesCache {
	cache := &SeriesCache{
		items: make(map[string]*CacheEntry),
	}
	// Start cleanup routine
	go cache.cleanup()
	return cache
}

// Get retrieves a cached series detail
func (c *SeriesCache) Get(key string) (*models.SeriesDetailResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Data, true
}

// Set stores a series detail in the cache
func (c *SeriesCache) Set(key string, data *models.SeriesDetailResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(CacheTTL),
	}
}

// cleanup removes expired entries every hour
func (c *SeriesCache) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.items {
			if now.After(entry.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Clear removes all cached entries
func (c *SeriesCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheEntry)
}

// Size returns the number of cached items
func (c *SeriesCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
