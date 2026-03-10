package domain

import (
	"time"

	"gorm.io/datatypes"
)

// Tenant represents a project/app (e.g., BUTL, Bship) in the multi-tenant CRM platform.
type Tenant struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	Metadata  datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// APIKey represents an authentication key issued to a tenant's project.
// There are two types: "server" (secret, for backend) and "client" (public, for app).
type APIKey struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	TenantID   uint64     `gorm:"index;not null" json:"tenant_id"`
	KeyHash    string     `gorm:"uniqueIndex;not null" json:"-"` // SHA-256 hash of the raw key
	KeyPrefix  string     `json:"key_prefix"`                    // "sk_live_" or "ck_live_"
	KeyType    string     `gorm:"not null" json:"key_type"`      // "server" or "client"
	Name       string     `json:"name"`                          // "Production Server", "Mobile Client"
	IsActive   bool       `gorm:"default:true" json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

// Key types
const (
	KeyTypeServer = "server"
	KeyTypeClient = "client"
)
