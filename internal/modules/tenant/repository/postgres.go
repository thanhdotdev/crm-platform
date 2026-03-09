package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/tenant/domain"
	"gorm.io/gorm"
)

// tenantPostgresRepo implements TenantRepository using GORM.
type tenantPostgresRepo struct {
	db *gorm.DB
}

// NewTenantPostgresRepo creates a new GORM-backed TenantRepository.
func NewTenantPostgresRepo(db *gorm.DB) TenantRepository {
	return &tenantPostgresRepo{db: db}
}

func (r *tenantPostgresRepo) Create(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantPostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &tenant, err
}

func (r *tenantPostgresRepo) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &tenant, err
}

func (r *tenantPostgresRepo) List(ctx context.Context, offset, limit int) ([]domain.Tenant, int64, error) {
	var tenants []domain.Tenant
	var total int64

	err := r.db.WithContext(ctx).Model(&domain.Tenant{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&tenants).Error
	return tenants, total, err
}

func (r *tenantPostgresRepo) Update(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *tenantPostgresRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Tenant{}).Error
}

// apiKeyPostgresRepo implements APIKeyRepository using GORM.
type apiKeyPostgresRepo struct {
	db *gorm.DB
}

// NewAPIKeyPostgresRepo creates a new GORM-backed APIKeyRepository.
func NewAPIKeyPostgresRepo(db *gorm.DB) APIKeyRepository {
	return &apiKeyPostgresRepo{db: db}
}

func (r *apiKeyPostgresRepo) Create(ctx context.Context, key *domain.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *apiKeyPostgresRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	var key domain.APIKey
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &key, err
}

func (r *apiKeyPostgresRepo) ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]domain.APIKey, error) {
	var keys []domain.APIKey
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *apiKeyPostgresRepo) Deactivate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&domain.APIKey{}).Where("id = ?", id).Update("is_active", false).Error
}
