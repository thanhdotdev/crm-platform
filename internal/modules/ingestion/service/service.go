package service

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/ingestion/domain"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// IngestionService handles business logic for ingestion.
type IngestionService interface {
	RecordEvent(ctx context.Context, event *domain.EventLog) error
}

type ingestionService struct {
	eventRepo domain.EventRepository
}

// NewIngestionService creates a new IngestionService.
func NewIngestionService(eventRepo domain.EventRepository) IngestionService {
	return &ingestionService{
		eventRepo: eventRepo,
	}
}

// RecordEvent stores a behavioral event in the timeline.
func (s *ingestionService) RecordEvent(ctx context.Context, event *domain.EventLog) error {
	if err := s.eventRepo.Create(ctx, event); err != nil {
		return apperror.Wrap("DB_ERROR", "failed to record event", err)
	}
	return nil
}
