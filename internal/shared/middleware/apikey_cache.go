package middleware

import (
	"sync"
	"time"
)

const apiKeyCacheTTL = 5 * time.Minute

// cachedKeyEntry holds cached API key lookup result.
type cachedKeyEntry struct {
	TenantID uint64
	KeyType  string
	CachedAt time.Time
}

// APIKeyCache provides in-memory caching for API key lookups.
// Uses sync.Map for concurrent safety, TTL-based expiration.
type APIKeyCache struct {
	store sync.Map // map[keyHash string] -> cachedKeyEntry
}

// NewAPIKeyCache creates a new API key cache.
func NewAPIKeyCache() *APIKeyCache {
	return &APIKeyCache{}
}

// Get returns cached entry if valid (not expired), or nil if miss/expired.
func (c *APIKeyCache) Get(keyHash string) *cachedKeyEntry {
	val, ok := c.store.Load(keyHash)
	if !ok {
		return nil
	}
	entry := val.(cachedKeyEntry)
	if time.Since(entry.CachedAt) > apiKeyCacheTTL {
		c.store.Delete(keyHash) // expired → evict
		return nil
	}
	return &entry
}

// Set stores an API key entry in the cache.
func (c *APIKeyCache) Set(keyHash string, tenantID uint64, keyType string) {
	c.store.Store(keyHash, cachedKeyEntry{
		TenantID: tenantID,
		KeyType:  keyType,
		CachedAt: time.Now(),
	})
}
