package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/trip/domain"
	"github.com/vothanh/crm-platform/internal/modules/trip/repository"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// TripService handles business logic for trip management.
type TripService struct {
	tripRepo repository.TripRepository
}

// NewTripService creates a new TripService.
func NewTripService(tripRepo repository.TripRepository) *TripService {
	return &TripService{tripRepo: tripRepo}
}

// CreateTrip creates a new trip record (from ingested event).
func (s *TripService) CreateTrip(ctx context.Context, trip *domain.Trip) (*domain.Trip, error) {
	// Check if trip with external ID already exists
	if trip.ExternalTripID != "" {
		existing, err := s.tripRepo.GetByExternalID(ctx, trip.TenantID, trip.ExternalTripID)
		if err != nil {
			return nil, apperror.Wrap("DB_ERROR", "failed to lookup trip", err)
		}
		if existing != nil {
			// Update existing trip
			existing.Status = trip.Status
			existing.Amount = trip.Amount
			existing.DiscountAmount = trip.DiscountAmount
			if trip.CompletedAt != nil {
				existing.CompletedAt = trip.CompletedAt
			}
			if trip.CancelReason != "" {
				existing.CancelReason = trip.CancelReason
			}
			if err := s.tripRepo.Update(ctx, existing); err != nil {
				return nil, apperror.Wrap("DB_ERROR", "failed to update trip", err)
			}
			return existing, nil
		}
	}

	if trip.BookedAt.IsZero() {
		trip.BookedAt = time.Now()
	}

	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create trip", err)
	}
	return trip, nil
}

// GetTrip retrieves a trip by ID.
func (s *TripService) GetTrip(ctx context.Context, tenantID, id uuid.UUID) (*domain.Trip, error) {
	trip, err := s.tripRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch trip", err)
	}
	if trip == nil {
		return nil, apperror.ErrNotFound
	}
	return trip, nil
}

// ListByCustomer returns trips for a specific customer.
func (s *TripService) ListByCustomer(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.Trip, int64, error) {
	return s.tripRepo.ListByCustomerID(ctx, tenantID, customerID, offset, limit)
}

// CompleteTrip marks a trip as completed.
func (s *TripService) CompleteTrip(ctx context.Context, tenantID, tripID uuid.UUID, amount float64) (*domain.Trip, error) {
	trip, err := s.tripRepo.GetByID(ctx, tenantID, tripID)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch trip", err)
	}
	if trip == nil {
		return nil, apperror.ErrNotFound
	}

	now := time.Now()
	trip.Status = domain.TripStatusCompleted
	trip.CompletedAt = &now
	if amount > 0 {
		trip.Amount = amount
	}

	if err := s.tripRepo.Update(ctx, trip); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to complete trip", err)
	}
	return trip, nil
}
