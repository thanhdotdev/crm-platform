package service

import (
	"context"
	"encoding/json"
	"time"

	customerDomain "github.com/vothanh/crm-platform/internal/modules/customer/domain"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/domain"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/repository"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// SegmentationService handles business logic for customer segmentation.
type SegmentationService struct {
	segmentRepo repository.SegmentRepository
	csRepo      repository.CustomerSegmentRepository
}

// NewSegmentationService creates a new SegmentationService.
func NewSegmentationService(
	segmentRepo repository.SegmentRepository,
	csRepo repository.CustomerSegmentRepository,
) *SegmentationService {
	return &SegmentationService{segmentRepo: segmentRepo, csRepo: csRepo}
}

// CreateSegment creates a new segment rule (per-tenant).
func (s *SegmentationService) CreateSegment(ctx context.Context, segment *domain.Segment) (*domain.Segment, error) {
	if err := s.segmentRepo.Create(ctx, segment); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create segment", err)
	}
	return segment, nil
}

// ListSegments returns all segments for a tenant.
func (s *SegmentationService) ListSegments(ctx context.Context, tenantID uint64) ([]domain.Segment, error) {
	return s.segmentRepo.List(ctx, tenantID)
}

// GetSegment returns a segment by ID (tenant-scoped).
func (s *SegmentationService) GetSegment(ctx context.Context, tenantID, id uint64) (*domain.Segment, error) {
	seg, err := s.segmentRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch segment", err)
	}
	if seg == nil {
		return nil, apperror.ErrNotFound
	}
	return seg, nil
}

// UpdateSegment updates a segment (per-tenant).
func (s *SegmentationService) UpdateSegment(ctx context.Context, segment *domain.Segment) (*domain.Segment, error) {
	if err := s.segmentRepo.Update(ctx, segment); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to update segment", err)
	}
	return segment, nil
}

// EvaluateCustomer checks if a customer matches a segment's rules and assigns them.
func (s *SegmentationService) EvaluateCustomer(ctx context.Context, tenantID uint64, customer *customerDomain.Customer) ([]uint64, error) {
	segments, err := s.segmentRepo.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var matchedSegmentIDs []uint64
	for _, seg := range segments {
		if !seg.IsActive {
			continue
		}
		if matchesRules(customer, json.RawMessage(seg.Rules)) {
			cs := &domain.CustomerSegment{
				TenantID:   tenantID,
				CustomerID: customer.ID,
				SegmentID:  seg.ID,
				Score:      customer.LeadScore,
				AssignedAt: time.Now(),
			}
			_ = s.csRepo.Assign(ctx, cs)
			matchedSegmentIDs = append(matchedSegmentIDs, seg.ID)
		}
	}

	return matchedSegmentIDs, nil
}

// GetSegmentCount returns the number of customers in a segment.
func (s *SegmentationService) GetSegmentCount(ctx context.Context, tenantID, segmentID uint64) (int64, error) {
	return s.csRepo.CountBySegmentID(ctx, tenantID, segmentID)
}

// matchesRules evaluates a customer against a segment's JSON rules.
func matchesRules(customer *customerDomain.Customer, rulesJSON json.RawMessage) bool {
	if len(rulesJSON) == 0 {
		return true
	}

	var rules domain.SegmentRules
	if err := json.Unmarshal(rulesJSON, &rules); err != nil {
		return false
	}

	if rules.MinTrips != nil && customer.TotalTrips < *rules.MinTrips {
		return false
	}
	if rules.MaxTrips != nil && customer.TotalTrips > *rules.MaxTrips {
		return false
	}
	if rules.MinSpent != nil && customer.TotalSpent < *rules.MinSpent {
		return false
	}
	if rules.MaxSpent != nil && customer.TotalSpent > *rules.MaxSpent {
		return false
	}
	if rules.MinLeadScore != nil && customer.LeadScore < *rules.MinLeadScore {
		return false
	}
	if rules.MaxLeadScore != nil && customer.LeadScore > *rules.MaxLeadScore {
		return false
	}
	if len(rules.LifecycleStages) > 0 {
		found := false
		for _, stage := range rules.LifecycleStages {
			if string(customer.LifecycleStage) == stage {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(rules.Sources) > 0 {
		found := false
		for _, src := range rules.Sources {
			if customer.Source == src {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
