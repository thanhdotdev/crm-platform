package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EventLog tracks user behavior events received from SDK.
type EventLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID  uuid.UUID      `gorm:"type:uuid;index;not null" json:"tenant_id"`
	UserID    string         `gorm:"index" json:"user_id"`
	UserType  string         `gorm:"index" json:"user_type"`
	EventType string         `gorm:"not null" json:"event_type"`
	EventData datatypes.JSON `gorm:"type:jsonb" json:"event_data,omitempty"`
	EventTime time.Time      `json:"event_time"`
	Source    string         `json:"source"` // "sdk", "api", "internal"
	CreatedAt time.Time      `json:"created_at"`
}

// BeforeCreate generates a UUID.
func (e *EventLog) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
