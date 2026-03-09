package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AutomationRule defines a per-tenant automation trigger + action.
type AutomationRule struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID        uuid.UUID      `gorm:"type:uuid;index;not null" json:"tenant_id"`
	Name            string         `gorm:"not null" json:"name"`
	TriggerType     string         `json:"trigger_type"` // event, schedule, condition
	TriggerConfig   datatypes.JSON `gorm:"type:jsonb" json:"trigger_config"`
	ActionType      string         `json:"action_type"` // send_notification, assign_voucher, update_segment
	ActionConfig    datatypes.JSON `gorm:"type:jsonb" json:"action_config"`
	TargetSegmentID *uuid.UUID     `gorm:"type:uuid" json:"target_segment_id,omitempty"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	ExecutionMode   string         `gorm:"default:'automatic'" json:"execution_mode"` // automatic, manual
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (a *AutomationRule) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AutomationLog tracks rule executions.
type AutomationLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"tenant_id"`
	RuleID     uuid.UUID      `gorm:"type:uuid;index" json:"rule_id"`
	CustomerID uuid.UUID      `gorm:"type:uuid;index" json:"customer_id"`
	Status     string         `json:"status"` // triggered, executed, failed, skipped, pending_approval
	Result     datatypes.JSON `gorm:"type:jsonb" json:"result,omitempty"`
	ExecutedAt time.Time      `json:"executed_at"`
}

func (al *AutomationLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}
