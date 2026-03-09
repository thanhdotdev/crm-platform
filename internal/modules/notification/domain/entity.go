package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Notification represents a notification sent to a customer.
type Notification struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"tenant_id"`
	CustomerID uuid.UUID  `gorm:"type:uuid;index" json:"customer_id"`
	Channel    string     `json:"channel"` // push, sms, email, zalo_oa, in_app
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Status     string     `gorm:"default:'pending'" json:"status"` // pending, sent, delivered, failed
	SentAt     *time.Time `json:"sent_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}

// NotificationProvider is the abstraction for sending notifications (SOLID - DIP).
// Each channel (push, sms, email, zalo) implements this interface.
type NotificationProvider interface {
	Send(ctx context.Context, notification *Notification) error
	Channel() string
}
