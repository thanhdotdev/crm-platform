package domain

import (
	"context"

	"github.com/google/uuid"
)

// CustomerService defines the business logic operations for customers.
type CustomerService interface {
	UpsertCustomer(ctx context.Context, customer *Customer) (*Customer, error)
	GetCustomer(ctx context.Context, tenantID, id uuid.UUID) (*Customer, error)
	ListCustomers(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]Customer, int64, error)
	IncrementTrip(ctx context.Context, tenantID, customerID uuid.UUID, amount float64) (bool, error)
	GetLifecycleCounts(ctx context.Context, tenantID uuid.UUID) (map[string]int64, error)
}
