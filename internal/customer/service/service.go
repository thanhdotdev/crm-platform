package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/customer/domain"
	"github.com/vothanh/crm-platform/internal/customer/repository"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// CustomerService handles business logic for customer management.
type CustomerService struct {
	customerRepo repository.CustomerRepository
	eventRepo    repository.CustomerEventRepository
}

// NewCustomerService creates a new CustomerService.
func NewCustomerService(
	customerRepo repository.CustomerRepository,
	eventRepo repository.CustomerEventRepository,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		eventRepo:    eventRepo,
	}
}

// UpsertCustomer creates or updates a customer (used by Ingestion API).
func (s *CustomerService) UpsertCustomer(ctx context.Context, customer *domain.Customer) (*domain.Customer, error) {
	if customer.ExternalID != "" {
		existing, err := s.customerRepo.GetByExternalID(ctx, customer.TenantID, customer.ExternalID)
		if err != nil {
			return nil, apperror.Wrap("DB_ERROR", "failed to lookup customer", err)
		}
		if existing != nil {
			existing.FullName = coalesce(customer.FullName, existing.FullName)
			existing.Email = coalesce(customer.Email, existing.Email)
			existing.Phone = coalesce(customer.Phone, existing.Phone)
			existing.Source = coalesce(customer.Source, existing.Source)
			existing.CampaignID = coalesce(customer.CampaignID, existing.CampaignID)
			existing.DeviceType = coalesce(customer.DeviceType, existing.DeviceType)
			if customer.AppInstalledAt != nil {
				existing.AppInstalledAt = customer.AppInstalledAt
			}
			if err := s.customerRepo.Update(ctx, existing); err != nil {
				return nil, apperror.Wrap("DB_ERROR", "failed to update customer", err)
			}
			return existing, nil
		}
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create customer", err)
	}
	return customer, nil
}

// GetCustomer retrieves a customer by ID (tenant-scoped).
func (s *CustomerService) GetCustomer(ctx context.Context, tenantID, id uuid.UUID) (*domain.Customer, error) {
	customer, err := s.customerRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch customer", err)
	}
	if customer == nil {
		return nil, apperror.ErrNotFound
	}
	return customer, nil
}

// ListCustomers returns a paginated list of customers (tenant-scoped).
func (s *CustomerService) ListCustomers(ctx context.Context, tenantID uuid.UUID, offset, limit int) ([]domain.Customer, int64, error) {
	return s.customerRepo.List(ctx, tenantID, offset, limit)
}

// RecordEvent stores a customer behavioral event and updates lead score.
func (s *CustomerService) RecordEvent(ctx context.Context, event *domain.CustomerEvent) error {
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return apperror.Wrap("DB_ERROR", "failed to record event", err)
	}

	scoreIncrement := getScoreForEvent(event.EventType)
	if scoreIncrement != 0 {
		customer, err := s.customerRepo.GetByID(ctx, event.TenantID, event.CustomerID)
		if err != nil || customer == nil {
			return nil
		}
		customer.LeadScore += scoreIncrement
		if customer.LeadScore < 0 {
			customer.LeadScore = 0
		}
		if customer.LeadScore > 100 {
			customer.LeadScore = 100
		}
		_ = s.customerRepo.Update(ctx, customer)
	}

	return nil
}

// IncrementTrip updates customer stats when a trip is completed.
// Returns true if customer was just upgraded to Luxury tier.
func (s *CustomerService) IncrementTrip(ctx context.Context, tenantID, customerID uuid.UUID, amount float64) (bool, error) {
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return false, apperror.Wrap("DB_ERROR", "failed to fetch customer", err)
	}
	if customer == nil {
		return false, apperror.ErrNotFound
	}

	customer.TotalTrips++
	customer.TotalSpent += amount
	now := time.Now()
	customer.LastTripAt = &now
	if customer.FirstTripAt == nil {
		customer.FirstTripAt = &now
	}

	customer.UpdateLifecycleStage()
	luxuryUpgraded := customer.CheckLuxuryEligibility()

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return false, apperror.Wrap("DB_ERROR", "failed to update customer stats", err)
	}

	return luxuryUpgraded, nil
}

// GetLifecycleCounts returns customer counts grouped by lifecycle stage.
func (s *CustomerService) GetLifecycleCounts(ctx context.Context, tenantID uuid.UUID) (map[string]int64, error) {
	return s.customerRepo.CountByLifecycle(ctx, tenantID)
}

// GetCustomerTimeline returns the event timeline for a customer.
func (s *CustomerService) GetCustomerTimeline(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.CustomerEvent, int64, error) {
	return s.eventRepo.ListByCustomerID(ctx, tenantID, customerID, offset, limit)
}

// Lead scoring rules per CEO requirements.
func getScoreForEvent(eventType string) int {
	scores := map[string]int{
		"user_registered": 10,
		"app_installed":   10,
		"app_opened":      5,
		"search_trip":     15,
		"enter_location":  10,
		"trip_booked":     20,
		"trip_completed":  50,
		"trip_cancelled":  -10,
	}
	return scores[eventType]
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
