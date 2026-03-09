package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/automation/domain"
	"gorm.io/gorm"
)

type automationRulePostgresRepo struct{ db *gorm.DB }

func NewAutomationRulePostgresRepo(db *gorm.DB) AutomationRuleRepository {
	return &automationRulePostgresRepo{db: db}
}

func (r *automationRulePostgresRepo) Create(ctx context.Context, rule *domain.AutomationRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *automationRulePostgresRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.AutomationRule, error) {
	var rule domain.AutomationRule
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&rule).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &rule, err
}

func (r *automationRulePostgresRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.AutomationRule, error) {
	var rules []domain.AutomationRule
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&rules).Error
	return rules, err
}

func (r *automationRulePostgresRepo) ListByTrigger(ctx context.Context, tenantID uuid.UUID, triggerType, eventType string) ([]domain.AutomationRule, error) {
	var rules []domain.AutomationRule
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND trigger_type = ? AND is_active = true", tenantID, triggerType)
	if eventType != "" {
		// Filter by event type in JSONB trigger_config
		query = query.Where("trigger_config->>'event' = ?", eventType)
	}
	err := query.Find(&rules).Error
	return rules, err
}

func (r *automationRulePostgresRepo) Update(ctx context.Context, rule *domain.AutomationRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

type automationLogPostgresRepo struct{ db *gorm.DB }

func NewAutomationLogPostgresRepo(db *gorm.DB) AutomationLogRepository {
	return &automationLogPostgresRepo{db: db}
}

func (r *automationLogPostgresRepo) Create(ctx context.Context, log *domain.AutomationLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *automationLogPostgresRepo) ListByRuleID(ctx context.Context, tenantID, ruleID uuid.UUID, offset, limit int) ([]domain.AutomationLog, int64, error) {
	var logs []domain.AutomationLog
	var total int64
	base := r.db.WithContext(ctx).Model(&domain.AutomationLog{}).Where("tenant_id = ? AND rule_id = ?", tenantID, ruleID)
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Order("executed_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *automationLogPostgresRepo) ListByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, offset, limit int) ([]domain.AutomationLog, int64, error) {
	var logs []domain.AutomationLog
	var total int64
	base := r.db.WithContext(ctx).Model(&domain.AutomationLog{}).Where("tenant_id = ? AND customer_id = ?", tenantID, customerID)
	base.Count(&total)
	err := base.Offset(offset).Limit(limit).Order("executed_at DESC").Find(&logs).Error
	return logs, total, err
}

// Ensure json import is used
var _ = json.Marshal
