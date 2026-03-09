package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SegmentType classifies how a segment is defined.
type SegmentType string

const (
	SegmentTypeTripBased  SegmentType = "trip_based"
	SegmentTypeLifecycle  SegmentType = "lifecycle"
	SegmentTypeCustom     SegmentType = "custom"
	SegmentTypeBehavioral SegmentType = "behavioral"
)

// Segment represents a customer grouping rule.
type Segment struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;index;not null" json:"tenant_id"`
	Name        string         `gorm:"not null" json:"name"`
	Type        SegmentType    `json:"type"`
	Rules       datatypes.JSON `gorm:"type:jsonb" json:"rules"`
	RiskLevel   string         `json:"risk_level,omitempty"`
	Description string         `json:"description,omitempty"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (s *Segment) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// CustomerSegment maps customers to segments.
type CustomerSegment struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID `gorm:"type:uuid;index;not null" json:"tenant_id"`
	CustomerID uuid.UUID `gorm:"type:uuid;index;not null" json:"customer_id"`
	SegmentID  uuid.UUID `gorm:"type:uuid;index;not null" json:"segment_id"`
	Score      int       `json:"score"`
	AssignedAt time.Time `json:"assigned_at"`
}

func (cs *CustomerSegment) BeforeCreate(tx *gorm.DB) error {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	return nil
}

// SegmentRules defines the JSON structure for segment evaluation.
type SegmentRules struct {
	MinTrips        *int     `json:"min_trips,omitempty"`
	MaxTrips        *int     `json:"max_trips,omitempty"`
	MinSpent        *float64 `json:"min_spent,omitempty"`
	MaxSpent        *float64 `json:"max_spent,omitempty"`
	MinLeadScore    *int     `json:"min_lead_score,omitempty"`
	MaxLeadScore    *int     `json:"max_lead_score,omitempty"`
	LifecycleStages []string `json:"lifecycle_stages,omitempty"`
	InactiveDays    *int     `json:"inactive_days,omitempty"`
	Sources         []string `json:"sources,omitempty"`
}
