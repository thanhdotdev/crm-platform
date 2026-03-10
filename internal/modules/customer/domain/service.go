package domain

import (
	"context"
)

// CustomerService defines the business logic operations for customers.
type CustomerService interface {
	UpsertCustomer(ctx context.Context, customer *Customer) (*Customer, error)
	GetCustomer(ctx context.Context, tenantID, id uint64) (*Customer, error)
	ListCustomers(ctx context.Context, tenantID uint64, offset, limit int) ([]Customer, int64, error)
	IncrementTrip(ctx context.Context, tenantID, customerID uint64, amount float64) (bool, error)
	GetLifecycleCounts(ctx context.Context, tenantID uint64) (map[string]int64, error)
}
