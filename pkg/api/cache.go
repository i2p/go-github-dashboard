package api

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-i2p/go-github-dashboard/pkg/types"
)

// CacheItem represents a cached item with its expiration time
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// Cache provides a simple caching mechanism for API responses
type Cache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
	dir   string
	ttl   time.Duration
}

// NewCache creates a new cache with the given directory and TTL
func NewCache(config *types.Config) *Cache {
	// Register types for gob encoding
	gob.Register([]types.Repository{})
	gob.Register([]types.PullRequest{})
	gob.Register([]types.Issue{})
	gob.Register([]types.Discussion{})

	cache := &Cache{
		items: make(map[string]CacheItem),
		dir:   config.CacheDir,
		ttl:   config.CacheTTL,
	}

	// Load cache from disk
	cache.loadCache()

	return cache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if time.Now().After(item.Expiration) {
		return nil, false
	}

	return item.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}

	// Save the cache to disk (in a separate goroutine to avoid blocking)
	go c.saveCache()
}

// Clear clears all items from the cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]CacheItem)
	go c.saveCache()
}

// saveCache saves the cache to disk
func (c *Cache) saveCache() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	cacheFile := filepath.Join(c.dir, "cache.gob")
	file, err := os.Create(cacheFile)
	if err != nil {
		fmt.Printf("Error creating cache file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(c.items)
	if err != nil {
		fmt.Printf("Error encoding cache: %v\n", err)
	}
}

// loadCache loads the cache from disk
func (c *Cache) loadCache() {
	cacheFile := filepath.Join(c.dir, "cache.gob")
	file, err := os.Open(cacheFile)
	if err != nil {
		// If the file doesn't exist, that's not an error
		if !os.IsNotExist(err) {
			fmt.Printf("Error opening cache file: %v\n", err)
		}
		return
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&c.items)
	if err != nil {
		fmt.Printf("Error decoding cache: %v\n", err)
		// If there's an error decoding, start with a fresh cache
		c.items = make(map[string]CacheItem)
	}

	// Remove expired items
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}
