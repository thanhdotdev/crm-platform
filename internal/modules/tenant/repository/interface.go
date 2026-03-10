package repository

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/tenant/domain"
)

// TenantRepository defines the interface for tenant data access.
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) error
	GetByID(ctx context.Context, id uint64) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context, offset, limit int) ([]domain.Tenant, int64, error)
	Update(ctx context.Context, tenant *domain.Tenant) error
	Delete(ctx context.Context, id uint64) error
}

// APIKeyRepository defines the interface for API key data access.
type APIKeyRepository interface {
	Create(ctx context.Context, key *domain.APIKey) error
	GetByID(ctx context.Context, id uint64) (*domain.APIKey, error)
	ListByTenantID(ctx context.Context, tenantID uint64) ([]domain.APIKey, error)
	Deactivate(ctx context.Context, id uint64) error
}
