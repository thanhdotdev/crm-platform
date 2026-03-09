package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/segmentation/domain"
	"gorm.io/gorm"
)

type segmentPostgresRepo struct{ db *gorm.DB }

func NewSegmentPostgresRepo(db *gorm.DB) SegmentRepository {
	return &segmentPostgresRepo{db: db}
}

func (r *segmentPostgresRepo) Create(ctx context.Context, s *domain.Segment) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *segmentPostgresRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Segment, error) {
	var s domain.Segment
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *segmentPostgresRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.Segment, error) {
	var segments []domain.Segment
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&segments).Error
	return segments, err
}

func (r *segmentPostgresRepo) Update(ctx context.Context, s *domain.Segment) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *segmentPostgresRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&domain.Segment{}).Error
}

type customerSegmentPostgresRepo struct{ db *gorm.DB }

func NewCustomerSegmentPostgresRepo(db *gorm.DB) CustomerSegmentRepository {
	return &customerSegmentPostgresRepo{db: db}
}

func (r *customerSegmentPostgresRepo) Assign(ctx context.Context, cs *domain.CustomerSegment) error {
	return r.db.WithContext(ctx).Create(cs).Error
}

func (r *customerSegmentPostgresRepo) Remove(ctx context.Context, tenantID, customerID, segmentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND customer_id = ? AND segment_id = ?", tenantID, customerID, segmentID).
		Delete(&domain.CustomerSegment{}).Error
}

func (r *customerSegmentPostgresRepo) ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) ([]domain.CustomerSegment, error) {
	var cs []domain.CustomerSegment
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID).Find(&cs).Error
	return cs, err
}

func (r *customerSegmentPostgresRepo) ListBySegmentID(ctx context.Context, tenantID, segmentID uuid.UUID, offset, limit int) ([]domain.CustomerSegment, int64, error) {
	var cs []domain.CustomerSegment
	var total int64
	base := r.db.WithContext(ctx).Model(&domain.CustomerSegment{}).Where("tenant_id = ? AND segment_id = ?", tenantID, segmentID)
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Find(&cs).Error
	return cs, total, err
}

func (r *customerSegmentPostgresRepo) CountBySegmentID(ctx context.Context, tenantID, segmentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.CustomerSegment{}).
		Where("tenant_id = ? AND segment_id = ?", tenantID, segmentID).Count(&count).Error
	return count, err
}
