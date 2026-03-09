package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/customer/domain"
)

// CustomerRepository defines the interface for customer data access.
type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Customer, error)
	GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Customer, error)
	GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.Customer, error)
	Upsert(ctx context.Context, customer *domain.Customer) error
	Update(ctx context.Context, customer *domain.Customer) error
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]domain.Customer, int64, error)
	CountByLifecycle(ctx context.Context, tenantID uuid.UUID) (map[string]int64, error)
}

// CustomerEventRepository defines the interface for customer event data access.
type CustomerEventRepository interface {
	Create(ctx context.Context, event *domain.EventLog) error
	ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.EventLog, int64, error)
}
