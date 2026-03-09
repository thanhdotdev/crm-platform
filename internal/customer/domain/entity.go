package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Customer is the core entity representing a customer in the CRM.
// It stores the 360° customer profile including identity, behavior metrics, and lifecycle state.
type Customer struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;index;not null" json:"tenant_id"`
	ExternalID     string         `gorm:"index" json:"external_id,omitempty"`
	Phone          string         `gorm:"not null" json:"phone"`
	FullName       string         `json:"full_name"`
	Email          string         `json:"email,omitempty"`
	Gender         string         `json:"gender,omitempty"`
	DateOfBirth    *time.Time     `json:"date_of_birth,omitempty"`
	AvatarURL      string         `json:"avatar_url,omitempty"`
	Source         string         `json:"source,omitempty"`
	CampaignID     string         `json:"campaign_id,omitempty"`
	DeviceType     string         `json:"device_type,omitempty"`
	LifecycleStage LifecycleStage `gorm:"default:'new_user'" json:"lifecycle_stage"`
	Tier           CustomerTier   `gorm:"default:'standard'" json:"tier"`
	LeadScore      int            `gorm:"default:0" json:"lead_score"`
	TotalTrips     int            `gorm:"default:0" json:"total_trips"` // completed trips only
	TotalSpent     float64        `gorm:"default:0" json:"total_spent"` // revenue from completed trips
	LastTripAt     *time.Time     `json:"last_trip_at,omitempty"`
	AppInstalledAt *time.Time     `json:"app_installed_at,omitempty"`
	FirstTripAt    *time.Time     `json:"first_trip_at,omitempty"`
	Metadata       datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// BeforeCreate generates a UUID before inserting.
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.LifecycleStage == "" {
		c.LifecycleStage = LifecycleNewUser
	}
	if c.Tier == "" {
		c.Tier = TierStandard
	}
	return nil
}

// UpdateLifecycleStage recalculates lifecycle based on completed trip count.
func (c *Customer) UpdateLifecycleStage() {
	switch {
	case c.TotalTrips >= 5:
		c.LifecycleStage = LifecycleVIP
	case c.TotalTrips >= 3:
		c.LifecycleStage = LifecycleLoyal
	case c.TotalTrips >= 2:
		c.LifecycleStage = LifecycleReturning
	case c.TotalTrips >= 1:
		c.LifecycleStage = LifecycleActivated
	case c.AppInstalledAt != nil:
		c.LifecycleStage = LifecycleInstalledNoTrip
	default:
		c.LifecycleStage = LifecycleNewUser
	}
}

// CheckLuxuryEligibility checks if the customer qualifies for Luxury tier.
// Returns true if the customer was just upgraded.
func (c *Customer) CheckLuxuryEligibility() bool {
	if c.Tier == TierLuxury {
		return false // already luxury
	}
	if c.TotalTrips >= LuxuryMinTrips && c.TotalSpent >= LuxuryMinSpent {
		c.Tier = TierLuxury
		return true // just upgraded
	}
	return false
}

// CustomerEvent tracks user behavior events received from SDK.
type CustomerEvent struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"tenant_id"`
	CustomerID uuid.UUID      `gorm:"type:uuid;index;not null" json:"customer_id"`
	EventType  string         `gorm:"not null" json:"event_type"`
	EventData  datatypes.JSON `gorm:"type:jsonb" json:"event_data,omitempty"`
	Source     string         `json:"source"` // "sdk", "api", "internal"
	CreatedAt  time.Time      `json:"created_at"`
}

// BeforeCreate generates a UUID.
func (e *CustomerEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
