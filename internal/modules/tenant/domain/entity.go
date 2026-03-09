package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Tenant represents a project/app (e.g., BUTL, Bship) in the multi-tenant CRM platform.
type Tenant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	Metadata  datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// BeforeCreate generates a UUID before inserting a new Tenant.
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// APIKey represents an authentication key issued to a tenant's project.
// There are two types: "server" (secret, for backend) and "client" (public, for app).
type APIKey struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID  `gorm:"index;not null" json:"tenant_id"`
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

// BeforeCreate generates a UUID before inserting a new APIKey.
func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// Key types
const (
	KeyTypeServer = "server"
	KeyTypeClient = "client"
)
