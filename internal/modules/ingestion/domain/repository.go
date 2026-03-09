package domain

import (
	"context"

	"github.com/google/uuid"
)

// EventRepository defines the interface for event data access.
type EventRepository interface {
	Create(ctx context.Context, event *EventLog) error
	ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]EventLog, int64, error)
}
