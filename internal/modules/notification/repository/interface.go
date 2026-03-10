package repository

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, tenantID, id uint64) (*domain.Notification, error)
	ListByCustomerID(ctx context.Context, tenantID, customerID uint64, offset, limit int) ([]domain.Notification, int64, error)
	UpdateStatus(ctx context.Context, id uint64, status string) error
}
