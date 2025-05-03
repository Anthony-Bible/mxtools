// Package cache provides caching mechanisms for DNS and blacklist results.
package cache

import (
	"sync"
	"time"
)

// Item represents a cached item.
type Item struct {
	Value      interface{}
	Expiration int64
}

// Cache is a simple in-memory cache with expiration.
type Cache struct {
	items map[string]Item
	mu    sync.RWMutex
}

// NewCache creates a new cache.
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]Item),
	}
	go cache.janitor()
	return cache
}

// Set adds an item to the cache with the specified expiration duration.
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(duration).UnixNano()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

// Get retrieves an item from the cache.
// The second return value indicates whether the key was found.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// Check if the item has expired
	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item)
}

// janitor periodically removes expired items from the cache.
func (c *Cache) janitor() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.deleteExpired()
	}
}

// deleteExpired deletes expired items from the cache.
func (c *Cache) deleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if now > v.Expiration {
			delete(c.items, k)
		}
	}
}

// DNSCache is a specialized cache for DNS results.
type DNSCache struct {
	cache *Cache
}

// NewDNSCache creates a new DNS cache.
func NewDNSCache() *DNSCache {
	return &DNSCache{
		cache: NewCache(),
	}
}

// Set adds a DNS result to the cache.
func (d *DNSCache) Set(domain string, recordType string, result interface{}) {
	key := domain + ":" + recordType
	d.cache.Set(key, result, 5*time.Minute) // Cache DNS results for 5 minutes
}

// Get retrieves a DNS result from the cache.
func (d *DNSCache) Get(domain string, recordType string) (interface{}, bool) {
	key := domain + ":" + recordType
	return d.cache.Get(key)
}

// BlacklistCache is a specialized cache for blacklist results.
type BlacklistCache struct {
	cache *Cache
}

// NewBlacklistCache creates a new blacklist cache.
func NewBlacklistCache() *BlacklistCache {
	return &BlacklistCache{
		cache: NewCache(),
	}
}

// Set adds a blacklist result to the cache.
func (b *BlacklistCache) Set(ip string, zone string, result interface{}) {
	key := ip + ":" + zone
	b.cache.Set(key, result, 30*time.Minute) // Cache blacklist results for 30 minutes
}

// Get retrieves a blacklist result from the cache.
func (b *BlacklistCache) Get(ip string, zone string) (interface{}, bool) {
	key := ip + ":" + zone
	return b.cache.Get(key)
}