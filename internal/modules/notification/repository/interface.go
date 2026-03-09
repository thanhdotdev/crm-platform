package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Notification, error)
	ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.Notification, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}
