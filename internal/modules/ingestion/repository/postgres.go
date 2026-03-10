package repository

import (
	"context"

	"github.com/vothanh/crm-platform/internal/modules/ingestion/domain"
	"gorm.io/gorm"
)

type eventPostgresRepo struct {
	db *gorm.DB
}

// NewEventPostgresRepo creates a new GORM-backed EventRepository.
func NewEventPostgresRepo(db *gorm.DB) domain.EventRepository {
	return &eventPostgresRepo{db: db}
}

func (r *eventPostgresRepo) Create(ctx context.Context, event *domain.EventLog) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *eventPostgresRepo) ListByCustomerID(ctx context.Context, tenantID, customerID uint64, offset, limit int) ([]domain.EventLog, int64, error) {
	var events []domain.EventLog
	var total int64

	base := r.db.WithContext(ctx).Model(&domain.EventLog{}).Where("tenant_id = ? AND user_id = ?", tenantID, customerID)
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Offset(offset).Limit(limit).Order("event_time DESC").Find(&events).Error
	return events, total, err
}
