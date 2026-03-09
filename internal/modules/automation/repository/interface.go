package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/automation/domain"
)

type AutomationRuleRepository interface {
	Create(ctx context.Context, rule *domain.AutomationRule) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.AutomationRule, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.AutomationRule, error)
	ListByTrigger(ctx context.Context, tenantID uuid.UUID, triggerType, eventType string) ([]domain.AutomationRule, error)
	Update(ctx context.Context, rule *domain.AutomationRule) error
}

type AutomationLogRepository interface {
	Create(ctx context.Context, log *domain.AutomationLog) error
	ListByRuleID(ctx context.Context, tenantID, ruleID uuid.UUID, offset, limit int) ([]domain.AutomationLog, int64, error)
	ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.AutomationLog, int64, error)
}
