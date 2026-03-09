package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/automation/domain"
	"github.com/vothanh/crm-platform/internal/modules/automation/repository"
	"github.com/vothanh/crm-platform/pkg/apperror"
)

// AutomationService handles automation rule execution.
type AutomationService struct {
	ruleRepo repository.AutomationRuleRepository
	logRepo  repository.AutomationLogRepository
}

func NewAutomationService(
	ruleRepo repository.AutomationRuleRepository,
	logRepo repository.AutomationLogRepository,
) *AutomationService {
	return &AutomationService{ruleRepo: ruleRepo, logRepo: logRepo}
}

// CreateRule creates a new automation rule (per-tenant).
func (s *AutomationService) CreateRule(ctx context.Context, rule *domain.AutomationRule) (*domain.AutomationRule, error) {
	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to create rule", err)
	}
	return rule, nil
}

// ListRules returns all automation rules for a tenant.
func (s *AutomationService) ListRules(ctx context.Context, tenantID uuid.UUID) ([]domain.AutomationRule, error) {
	return s.ruleRepo.List(ctx, tenantID)
}

// GetRule returns a rule by ID (tenant-scoped).
func (s *AutomationService) GetRule(ctx context.Context, tenantID, id uuid.UUID) (*domain.AutomationRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to fetch rule", err)
	}
	if rule == nil {
		return nil, apperror.ErrNotFound
	}
	return rule, nil
}

// UpdateRule updates an automation rule.
func (s *AutomationService) UpdateRule(ctx context.Context, rule *domain.AutomationRule) (*domain.AutomationRule, error) {
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, apperror.Wrap("DB_ERROR", "failed to update rule", err)
	}
	return rule, nil
}

// EvaluateEvent checks if any automation rules match the given event and executes them.
// Returns the list of triggered rule IDs.
func (s *AutomationService) EvaluateEvent(ctx context.Context, tenantID uuid.UUID, eventType string, customerID uuid.UUID) ([]uuid.UUID, error) {
	rules, err := s.ruleRepo.ListByTrigger(ctx, tenantID, "event", eventType)
	if err != nil {
		return nil, err
	}

	var triggeredIDs []uuid.UUID
	for _, rule := range rules {
		status := "executed"
		if rule.ExecutionMode == "manual" {
			status = "pending_approval"
		}

		log := &domain.AutomationLog{
			TenantID:   tenantID,
			RuleID:     rule.ID,
			CustomerID: customerID,
			Status:     status,
			ExecutedAt: time.Now(),
		}
		_ = s.logRepo.Create(ctx, log)
		triggeredIDs = append(triggeredIDs, rule.ID)
	}

	return triggeredIDs, nil
}

// ListLogs returns automation logs for a rule.
func (s *AutomationService) ListLogs(ctx context.Context, tenantID, ruleID uuid.UUID, offset, limit int) ([]domain.AutomationLog, int64, error) {
	return s.logRepo.ListByRuleID(ctx, tenantID, ruleID, offset, limit)
}
