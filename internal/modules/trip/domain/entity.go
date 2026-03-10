package domain

import (
	"time"
)

// TripStatus represents the status of a trip.
type TripStatus string

const (
	TripStatusBooked     TripStatus = "booked"
	TripStatusInProgress TripStatus = "in_progress"
	TripStatusCompleted  TripStatus = "completed"
	TripStatusCancelled  TripStatus = "cancelled"
)

// Trip represents a ride/service request in the CRM.
type Trip struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	TenantID        uint64     `gorm:"index;not null" json:"tenant_id"`
	CustomerID      uint64     `gorm:"index;not null" json:"customer_id"`
	ExternalTripID  string     `gorm:"index" json:"external_trip_id,omitempty"`
	Status          TripStatus `gorm:"default:'booked'" json:"status"`
	PickupLocation  string     `json:"pickup_location,omitempty"`
	DropoffLocation string     `json:"dropoff_location,omitempty"`
	PickupLat       float64    `json:"pickup_lat,omitempty"`
	PickupLng       float64    `json:"pickup_lng,omitempty"`
	DropoffLat      float64    `json:"dropoff_lat,omitempty"`
	DropoffLng      float64    `json:"dropoff_lng,omitempty"`
	Amount          float64    `json:"amount"`
	DiscountAmount  float64    `json:"discount_amount"`
	VoucherID       *uint64    `json:"voucher_id,omitempty"`
	CancelReason    string     `json:"cancel_reason,omitempty"`
	BookedAt        time.Time  `json:"booked_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
