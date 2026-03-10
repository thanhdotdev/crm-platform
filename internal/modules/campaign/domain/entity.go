package domain

import (
	"time"
)

// Campaign represents a marketing campaign.
type Campaign struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	TenantID        uint64    `gorm:"index;not null" json:"tenant_id"`
	Name            string    `gorm:"not null" json:"name"`
	Type            string    `json:"type"`                          // new_user, reactivation, loyalty, luxury_upgrade
	Status          string    `gorm:"default:'draft'" json:"status"` // draft, active, paused, completed
	TargetSegmentID *uint64   `json:"target_segment_id,omitempty"`
	Budget          float64   `json:"budget"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Description     string    `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Voucher represents a discount voucher within a campaign.
type Voucher struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	TenantID      uint64    `gorm:"index;not null" json:"tenant_id"`
	CampaignID    uint64    `gorm:"index" json:"campaign_id"`
	Code          string    `gorm:"uniqueIndex;not null" json:"code"`
	DiscountType  string    `json:"discount_type"`  // fixed, percentage
	DiscountValue float64   `json:"discount_value"` // 50000, 30000, 20000 (VND)
	MinOrderValue float64   `json:"min_order_value"`
	MaxUsageCount int       `json:"max_usage_count"`
	UsedCount     int       `gorm:"default:0" json:"used_count"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsStackable   bool      `gorm:"default:false" json:"is_stackable"`  // false per CEO policy
	PeakHourOnly  bool      `gorm:"default:true" json:"peak_hour_only"` // true — allowed during peak
	TripNumber    int       `json:"trip_number"`                        // which trip this voucher is for (1, 2, 3)
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

// VoucherUsage tracks when a voucher is redeemed.
type VoucherUsage struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	TenantID   uint64    `gorm:"index;not null" json:"tenant_id"`
	VoucherID  uint64    `gorm:"index" json:"voucher_id"`
	CustomerID uint64    `gorm:"index" json:"customer_id"`
	TripID     uint64    `json:"trip_id"`
	UsedAt     time.Time `json:"used_at"`
}

// Policy50_30_20 returns the discount value for a given trip number (CEO requirement).
func Policy50_30_20(tripNumber int) (discountValue float64, expiryHours int) {
	switch tripNumber {
	case 1:
		return 50000, 48 // 50K VND, 24-48h urgency
	case 2:
		return 30000, 720 // 30K VND, 14-30 days
	case 3:
		return 20000, 720 // 20K VND, 14-30 days
	default:
		return 0, 0
	}
}
