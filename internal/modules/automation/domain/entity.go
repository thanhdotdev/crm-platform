package domain

import (
	"time"

	"gorm.io/datatypes"
)

// AutomationRule defines a per-tenant automation trigger + action.
type AutomationRule struct {
	ID              uint64         `gorm:"primaryKey" json:"id"`
	TenantID        uint64         `gorm:"index;not null" json:"tenant_id"`
	Name            string         `gorm:"not null" json:"name"`
	TriggerType     string         `json:"trigger_type"` // event, schedule, condition
	TriggerConfig   datatypes.JSON `gorm:"type:jsonb" json:"trigger_config"`
	ActionType      string         `json:"action_type"` // send_notification, assign_voucher, update_segment
	ActionConfig    datatypes.JSON `gorm:"type:jsonb" json:"action_config"`
	TargetSegmentID *uint64        `json:"target_segment_id,omitempty"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	ExecutionMode   string         `gorm:"default:'automatic'" json:"execution_mode"` // automatic, manual
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// AutomationLog tracks rule executions.
type AutomationLog struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	TenantID   uint64         `gorm:"index;not null" json:"tenant_id"`
	RuleID     uint64         `gorm:"index" json:"rule_id"`
	CustomerID uint64         `gorm:"index" json:"customer_id"`
	Status     string         `json:"status"` // triggered, executed, failed, skipped, pending_approval
	Result     datatypes.JSON `gorm:"type:jsonb" json:"result,omitempty"`
	ExecutedAt time.Time      `json:"executed_at"`
}
