package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/trip/domain"
	"gorm.io/gorm"
)

type tripPostgresRepo struct {
	db *gorm.DB
}

// NewTripPostgresRepo creates a new GORM-backed TripRepository.
func NewTripPostgresRepo(db *gorm.DB) TripRepository {
	return &tripPostgresRepo{db: db}
}

func (r *tripPostgresRepo) Create(ctx context.Context, trip *domain.Trip) error {
	return r.db.WithContext(ctx).Create(trip).Error
}

func (r *tripPostgresRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Trip, error) {
	var trip domain.Trip
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&trip).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &trip, err
}

func (r *tripPostgresRepo) GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalTripID string) (*domain.Trip, error) {
	var trip domain.Trip
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND external_trip_id = ?", tenantID, externalTripID).First(&trip).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &trip, err
}

func (r *tripPostgresRepo) ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.Trip, int64, error) {
	var trips []domain.Trip
	var total int64

	base := r.db.WithContext(ctx).Model(&domain.Trip{}).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&trips).Error
	return trips, total, err
}

func (r *tripPostgresRepo) Update(ctx context.Context, trip *domain.Trip) error {
	return r.db.WithContext(ctx).Save(trip).Error
}

// CountCompletedByCustomerID counts only COMPLETED trips (for lifecycle/tier logic).
func (r *tripPostgresRepo) CountCompletedByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Trip{}).
		Where("tenant_id = ? AND customer_id = ? AND status = ?", tenantID, customerID, domain.TripStatusCompleted).
		Count(&count).Error
	return count, err
}
