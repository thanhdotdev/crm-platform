package domain

import (
	"time"

	"gorm.io/datatypes"
)

// EventLog tracks user behavior events received from SDK.
type EventLog struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	TenantID  uint64         `gorm:"index;not null" json:"tenant_id"`
	UserID    string         `gorm:"index" json:"user_id"`
	UserType  string         `gorm:"index" json:"user_type"`
	EventType string         `gorm:"not null" json:"event_type"`
	EventData datatypes.JSON `gorm:"type:jsonb" json:"event_data,omitempty"`
	EventTime time.Time      `json:"event_time"`
	Source    string         `json:"source"` // "sdk", "api", "internal"
	CreatedAt time.Time      `json:"created_at"`
}
