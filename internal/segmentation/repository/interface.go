package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/segmentation/domain"
)

// SegmentRepository defines the interface for segment data access.
type SegmentRepository interface {
	Create(ctx context.Context, segment *domain.Segment) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Segment, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Segment, error)
	Update(ctx context.Context, segment *domain.Segment) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
}

// CustomerSegmentRepository defines the interface for customer-segment mapping.
type CustomerSegmentRepository interface {
	Assign(ctx context.Context, cs *domain.CustomerSegment) error
	Remove(ctx context.Context, tenantID, customerID, segmentID uuid.UUID) error
	ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) ([]domain.CustomerSegment, error)
	ListBySegmentID(ctx context.Context, tenantID, segmentID uuid.UUID, offset, limit int) ([]domain.CustomerSegment, int64, error)
	CountBySegmentID(ctx context.Context, tenantID, segmentID uuid.UUID) (int64, error)
}
