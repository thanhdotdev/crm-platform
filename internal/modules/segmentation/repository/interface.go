package repository

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/segmentation/domain"
)

// SegmentRepository defines the interface for segment data access.
type SegmentRepository interface {
	Create(ctx context.Context, segment *domain.Segment) error
	GetByID(ctx context.Context, tenantID, id uint64) (*domain.Segment, error)
	List(ctx context.Context, tenantID uint64) ([]domain.Segment, error)
	Update(ctx context.Context, segment *domain.Segment) error
	Delete(ctx context.Context, tenantID, id uint64) error
}

// CustomerSegmentRepository defines the interface for customer-segment mapping.
type CustomerSegmentRepository interface {
	Assign(ctx context.Context, cs *domain.CustomerSegment) error
	Remove(ctx context.Context, tenantID, customerID, segmentID uint64) error
	ListByCustomerID(ctx context.Context, tenantID, customerID uint64) ([]domain.CustomerSegment, error)
	ListBySegmentID(ctx context.Context, tenantID, segmentID uint64, offset, limit int) ([]domain.CustomerSegment, int64, error)
	CountBySegmentID(ctx context.Context, tenantID, segmentID uint64) (int64, error)
}
