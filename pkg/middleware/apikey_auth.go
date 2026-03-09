package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/pkg/response"
	"gorm.io/gorm"
)

// contextKey constants for tenant resolution.
const (
	ContextKeyTenantID = "tenant_id"
	ContextKeyKeyType  = "key_type"
)

// apiKeyRecord represents the database row for API key lookup.
type apiKeyRecord struct {
	TenantID uuid.UUID
	KeyType  string
	IsActive bool
}

// tenantRecord represents the database row for tenant lookup.
type tenantRecord struct {
	IsActive bool
}

// APIKeyAuth returns a middleware that authenticates requests via X-API-Key header.
// Uses in-memory cache to avoid DB lookups on every request.
func APIKeyAuth(db *gorm.DB, cache *APIKeyCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			response.Unauthorized(c, "missing X-API-Key header")
			c.Abort()
			return
		}

		keyHash := hashAPIKey(apiKey)

		// Check cache first → 0 DB queries on hit
		if entry := cache.Get(keyHash); entry != nil {
			c.Set(ContextKeyTenantID, entry.TenantID)
			c.Set(ContextKeyKeyType, entry.KeyType)
			c.Next()
			return
		}

		// Cache miss → query DB
		var record apiKeyRecord
		err := db.Table("api_keys").
			Select("tenant_id, key_type, is_active").
			Where("key_hash = ?", keyHash).
			First(&record).Error

		if err != nil {
			response.Unauthorized(c, "invalid API key")
			c.Abort()
			return
		}

		if !record.IsActive {
			response.Unauthorized(c, "API key is inactive")
			c.Abort()
			return
		}

		// Verify tenant is active
		var tenant tenantRecord
		err = db.Table("tenants").
			Select("is_active").
			Where("id = ?", record.TenantID).
			First(&tenant).Error

		if err != nil || !tenant.IsActive {
			response.Unauthorized(c, "tenant is inactive")
			c.Abort()
			return
		}

		// Cache the valid key
		cache.Set(keyHash, record.TenantID, record.KeyType)

		// Update last_used_at (async, non-blocking)
		go func() {
			db.Table("api_keys").Where("key_hash = ?", keyHash).
				Update("last_used_at", gorm.Expr("NOW()"))
		}()

		// Inject tenant info into context
		c.Set(ContextKeyTenantID, record.TenantID)
		c.Set(ContextKeyKeyType, record.KeyType)

		c.Next()
	}
}

// GetTenantID extracts the TenantID from gin.Context.
func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyTenantID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetKeyType extracts the key type (server/client) from gin.Context.
func GetKeyType(c *gin.Context) string {
	val, _ := c.Get(ContextKeyKeyType)
	keyType, _ := val.(string)
	return keyType
}

// hashAPIKey creates a SHA-256 hash of the raw API key.
func hashAPIKey(key string) string {
	h := sha256.New()
	h.Write([]byte(strings.TrimSpace(key)))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateAPIKeyHash is a utility to hash a raw key for storage.
func GenerateAPIKeyHash(rawKey string) string {
	return hashAPIKey(rawKey)
}
