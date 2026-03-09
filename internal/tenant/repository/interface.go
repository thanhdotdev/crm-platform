package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/tenant/domain"
)

// TenantRepository defines the interface for tenant data access.
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]domain.Tenant, int64, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// APIKeyRepository defines the interface for API key data access.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]domain.APIKey, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
}
