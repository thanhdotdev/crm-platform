package domain

import (
	"context"
)

// CustomerRepository defines the interface for customer data access.
type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, tenantID, id uint64) (*Customer, error)
	GetByExternalID(ctx context.Context, tenantID uint64, externalID string) (*Customer, error)
	GetByPhone(ctx context.Context, tenantID uint64, phone string) (*Customer, error)
	Upsert(ctx context.Context, customer *Customer) error
	Update(ctx context.Context, customer *Customer) error
	List(ctx context.Context, tenantID uint64, offset, limit int) ([]Customer, int64, error)
	CountByLifecycle(ctx context.Context, tenantID uint64) (map[string]int64, error)
}
