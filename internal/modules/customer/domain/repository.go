package domain

import (
	"context"

	"github.com/google/uuid"
)

// CustomerRepository defines the interface for customer data access.
type CustomerRepository interface {
	Create(ctx context.Context, customer *Customer) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*Customer, error)
	GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*Customer, error)
	GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*Customer, error)
	Upsert(ctx context.Context, customer *Customer) error
	Update(ctx context.Context, customer *Customer) error
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]Customer, int64, error)
	CountByLifecycle(ctx context.Context, tenantID uuid.UUID) (map[string]int64, error)
}
