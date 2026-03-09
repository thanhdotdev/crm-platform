package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/customer/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type customerPostgresRepo struct {
	db *gorm.DB
}

// NewCustomerPostgresRepo creates a new GORM-backed CustomerRepository.
func NewCustomerPostgresRepo(db *gorm.DB) CustomerRepository {
	return &customerPostgresRepo{db: db}
}

func (r *customerPostgresRepo) Create(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *customerPostgresRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&customer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &customer, err
}

func (r *customerPostgresRepo) GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND external_id = ?", tenantID, externalID).First(&customer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &customer, err
}

func (r *customerPostgresRepo) GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND phone = ?", tenantID, phone).First(&customer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &customer, err
}

// Upsert creates or updates a customer based on (tenant_id, external_id).
func (r *customerPostgresRepo) Upsert(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "external_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"full_name", "email", "phone", "source", "campaign_id", "device_type", "metadata", "updated_at"}),
		}).
		Create(customer).Error
}

func (r *customerPostgresRepo) Update(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *customerPostgresRepo) List(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
	var total int64

	base := r.db.WithContext(ctx).Model(&domain.Customer{}).Where("tenant_id = ?", tenantID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&customers).Error
	return customers, total, err
}

func (r *customerPostgresRepo) CountByLifecycle(ctx context.Context, tenantID uuid.UUID) (map[string]int64, error) {
	type result struct {
		LifecycleStage string
		Count          int64
	}
	var results []result
	err := r.db.WithContext(ctx).
		Model(&domain.Customer{}).
		Select("lifecycle_stage, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("lifecycle_stage").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	m := make(map[string]int64)
	for _, r := range results {
		m[r.LifecycleStage] = r.Count
	}
	return m, nil
}

// customerEventPostgresRepo implements CustomerEventRepository.
type customerEventPostgresRepo struct {
	db *gorm.DB
}

// NewCustomerEventPostgresRepo creates a new GORM-backed CustomerEventRepository.
func NewCustomerEventPostgresRepo(db *gorm.DB) CustomerEventRepository {
	return &customerEventPostgresRepo{db: db}
}

func (r *customerEventPostgresRepo) Create(ctx context.Context, event *domain.EventLog) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *customerEventPostgresRepo) ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.EventLog, int64, error) {
	var events []domain.EventLog
	var total int64

	base := r.db.WithContext(ctx).Model(&domain.EventLog{}).Where("tenant_id = ? AND user_id = ?", tenantID, customerID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Offset(offset).Limit(limit).Order("event_time DESC").Find(&events).Error
	return events, total, err
}
