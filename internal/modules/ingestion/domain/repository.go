package domain

import (
	"context"
)

// EventRepository defines the interface for event data access.
type EventRepository interface {
	Create(ctx context.Context, event *EventLog) error
	ListByCustomerID(ctx context.Context, tenantID, customerID uint64, offset, limit int) ([]EventLog, int64, error)
}
