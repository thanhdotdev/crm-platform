package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/tenant/domain"
	"github.com/vothanh/crm-platform/internal/modules/tenant/repository"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// TenantService handles business logic for tenant management.
type TenantService struct {
	tenantRepo repository.TenantRepository
	apiKeyRepo repository.APIKeyRepository
}

// NewTenantService creates a new TenantService.
func NewTenantService(tenantRepo repository.TenantRepository, apiKeyRepo repository.APIKeyRepository) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		apiKeyRepo: apiKeyRepo,
	}
}

// CreateTenant creates a new tenant with a unique slug.
func (s *TenantService) CreateTenant(ctx context.Context, name, slug string) (*domain.Tenant, error) {
	existing, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to check slug uniqueness", err)
	}
	if existing != nil {
		return nil, apperror.New("SLUG_EXISTS", fmt.Sprintf("tenant with slug '%s' already exists", slug))
	}

	tenant := &domain.Tenant{
		Name:     name,
		Slug:     slug,
		IsActive: true,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create tenant", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID.
func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch tenant", err)
	}
	if tenant == nil {
		return nil, apperror.ErrNotFound
	}
	return tenant, nil
}

// ListTenants returns a paginated list of tenants.
func (s *TenantService) ListTenants(ctx context.Context, offset, limit int) ([]domain.Tenant, int64, error) {
	return s.tenantRepo.List(ctx, offset, limit)
}

// UpdateTenant updates a tenant's mutable fields.
func (s *TenantService) UpdateTenant(ctx context.Context, id uuid.UUID, name string, isActive *bool) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch tenant", err)
	}
	if tenant == nil {
		return nil, apperror.ErrNotFound
	}

	if name != "" {
		tenant.Name = name
	}
	if isActive != nil {
		tenant.IsActive = *isActive
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to update tenant", err)
	}

	return tenant, nil
}

// APIKeyWithRaw wraps APIKey with the raw key (only returned on creation).
type APIKeyWithRaw struct {
	domain.APIKey
	RawKey string `json:"raw_key"`
}

// CreateAPIKey generates a new API key for a tenant.
func (s *TenantService) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, keyType, name string) (*APIKeyWithRaw, error) {
	// Validate tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch tenant", err)
	}
	if tenant == nil {
		return nil, apperror.ErrNotFound
	}

	// Validate key type
	if keyType != domain.KeyTypeServer && keyType != domain.KeyTypeClient {
		return nil, apperror.New("INVALID_KEY_TYPE", "key type must be 'server' or 'client'")
	}

	// Generate raw key
	prefix := "sk_live_"
	if keyType == domain.KeyTypeClient {
		prefix = "ck_live_"
	}
	rawKey := prefix + generateRandomKey(32)
	keyHash := middleware.GenerateAPIKeyHash(rawKey)

	apiKey := &domain.APIKey{
		TenantID:  tenantID,
		KeyHash:   keyHash,
		KeyPrefix: prefix,
		KeyType:   keyType,
		Name:      name,
		IsActive:  true,
	}

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create API key", err)
	}

	return &APIKeyWithRaw{
		APIKey: *apiKey,
		RawKey: rawKey,
	}, nil
}

// ListAPIKeys returns all API keys for a tenant.
func (s *TenantService) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]domain.APIKey, error) {
	return s.apiKeyRepo.ListByTenantID(ctx, tenantID)
}

// DeactivateAPIKey disables an API key.
func (s *TenantService) DeactivateAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.apiKeyRepo.Deactivate(ctx, keyID)
}

// generateRandomKey creates a cryptographically random hex string.
func generateRandomKey(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
