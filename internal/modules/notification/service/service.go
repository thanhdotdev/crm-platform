package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
	"github.com/vothanh/crm-platform/internal/modules/notification/repository"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// NotificationService handles multi-channel notification delivery.
type NotificationService struct {
	repo      repository.NotificationRepository
	providers map[string]domain.NotificationProvider
}

func NewNotificationService(
	repo repository.NotificationRepository,
	providers []domain.NotificationProvider,
) *NotificationService {
	pm := make(map[string]domain.NotificationProvider)
	for _, p := range providers {
		pm[p.Channel()] = p
	}
	return &NotificationService{repo: repo, providers: pm}
}

// Send creates and sends a notification via the appropriate channel provider.
func (s *NotificationService) Send(ctx context.Context, n *domain.Notification) (*domain.Notification, error) {
	provider, ok := s.providers[n.Channel]
	if !ok {
		return nil, apperror.New("UNSUPPORTED_CHANNEL", fmt.Sprintf("channel '%s' is not supported", n.Channel))
	}

	// Save notification record
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to save notification", err)
	}

	// Send via provider
	if err := provider.Send(ctx, n); err != nil {
		_ = s.repo.UpdateStatus(ctx, n.ID, "failed")
		return n, apperror.Wrap("SEND_FAILED", "failed to send notification", err)
	}

	now := time.Now()
	n.Status = "sent"
	n.SentAt = &now
	_ = s.repo.UpdateStatus(ctx, n.ID, "sent")

	return n, nil
}

// ListByCustomer returns notifications for a customer.
func (s *NotificationService) ListByCustomer(ctx context.Context, tenantID, customerID uint64, offset, limit int) ([]domain.Notification, int64, error) {
	return s.repo.ListByCustomerID(ctx, tenantID, customerID, offset, limit)
}

// AvailableChannels returns the registered channels.
func (s *NotificationService) AvailableChannels() []string {
	channels := make([]string, 0, len(s.providers))
	for ch := range s.providers {
		channels = append(channels, ch)
	}
	return channels
}
