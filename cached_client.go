package gocloudclient

import (
	"fmt"
	"sync"
	"time"
)

// CachedClient wraps a Client with caching capabilities
type CachedClient struct {
	client     *Client
	cache      map[string]*cacheEntry
	cacheMutex sync.RWMutex
	defaultTTL time.Duration
}

type cacheEntry struct {
	config    *ConfigResponse
	expiresAt time.Time
}

// NewCachedClient creates a new cached client wrapper
func NewCachedClient(client *Client, defaultTTL time.Duration) *CachedClient {
	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute // Default 5 minutes
	}

	return &CachedClient{
		client:     client,
		cache:      make(map[string]*cacheEntry),
		defaultTTL: defaultTTL,
	}
}

// GetConfig fetches configuration with caching
func (c *CachedClient) GetConfig(application, profile, label string) (*ConfigResponse, error) {
	cacheKey := fmt.Sprintf("%s:%s:%s", application, profile, label)

	// Try to get from cache
	c.cacheMutex.RLock()
	if entry, exists := c.cache[cacheKey]; exists {
		if time.Now().Before(entry.expiresAt) {
			config := entry.config
			c.cacheMutex.RUnlock()
			return config, nil
		}
	}
	c.cacheMutex.RUnlock()

	// Fetch from server
	config, err := c.client.GetConfig()
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.cacheMutex.Lock()
	c.cache[cacheKey] = &cacheEntry{
		config:    config,
		expiresAt: time.Now().Add(c.defaultTTL),
	}
	c.cacheMutex.Unlock()

	return config, nil
}

// ClearCache clears all cached entries
func (c *CachedClient) ClearCache() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.cache = make(map[string]*cacheEntry)
}

// InvalidateCache invalidates cache for a specific application/profile/label combination
func (c *CachedClient) InvalidateCache(application, profile, label string) {
	cacheKey := fmt.Sprintf("%s:%s:%s", application, profile, label)
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.cache, cacheKey)
}
