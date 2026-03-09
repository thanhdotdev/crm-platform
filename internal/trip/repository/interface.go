package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/trip/domain"
)

// TripRepository defines the interface for trip data access.
type TripRepository interface {
	Create(ctx context.Context, trip *domain.Trip) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Trip, error)
	GetByExternalID(ctx context.Context, tenantID uuid.UUID, externalTripID string) (*domain.Trip, error)
	ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.Trip, int64, error)
	Update(ctx context.Context, trip *domain.Trip) error
	CountCompletedByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) (int64, error)
}
