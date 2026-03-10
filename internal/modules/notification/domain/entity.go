package domain

import (
	"context"
	"time"
)

// Notification represents a notification sent to a customer.
type Notification struct {
	ID         uint64     `gorm:"primaryKey" json:"id"`
	TenantID   uint64     `gorm:"index;not null" json:"tenant_id"`
	CustomerID uint64     `gorm:"index" json:"customer_id"`
	Channel    string     `json:"channel"` // push, sms, email, zalo_oa, in_app
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Status     string     `gorm:"default:'pending'" json:"status"` // pending, sent, delivered, failed
	SentAt     *time.Time `json:"sent_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// NotificationProvider is the abstraction for sending notifications (SOLID - DIP).
// Each channel (push, sms, email, zalo) implements this interface.
type NotificationProvider interface {
	Send(ctx context.Context, notification *Notification) error
	Channel() string
}
