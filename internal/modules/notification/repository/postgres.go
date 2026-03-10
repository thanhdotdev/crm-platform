package repository

import (
	"context"
	"errors"

	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
	"gorm.io/gorm"
)

type notificationPostgresRepo struct{ db *gorm.DB }

func NewNotificationPostgresRepo(db *gorm.DB) NotificationRepository {
	return &notificationPostgresRepo{db: db}
}

func (r *notificationPostgresRepo) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationPostgresRepo) GetByID(ctx context.Context, tenantID, id uint64) (*domain.Notification, error) {
	var n domain.Notification
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &n, err
}

func (r *notificationPostgresRepo) ListByCustomerID(ctx context.Context, tenantID, customerID uint64, offset, limit int) ([]domain.Notification, int64, error) {
	var notifications []domain.Notification
	var total int64
	base := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID)
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error
	return notifications, total, err
}

func (r *notificationPostgresRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", id).Update("status", status).Error
}
