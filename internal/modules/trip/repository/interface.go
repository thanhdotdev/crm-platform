package repository

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/trip/domain"
)

// TripRepository defines the interface for trip data access.
type TripRepository interface {
	Create(ctx context.Context, trip *domain.Trip) error
	GetByID(ctx context.Context, tenantID, id uint64) (*domain.Trip, error)
	GetByExternalID(ctx context.Context, tenantID uint64, externalTripID string) (*domain.Trip, error)
	ListByCustomerID(ctx context.Context, tenantID, customerID uint64, offset, limit int) ([]domain.Trip, int64, error)
	Update(ctx context.Context, trip *domain.Trip) error
	CountCompletedByCustomerID(ctx context.Context, tenantID, customerID uint64) (int64, error)
}
