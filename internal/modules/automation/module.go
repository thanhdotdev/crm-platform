package automation

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/automation/handler"
	"github.com/vothanh/crm-platform/internal/modules/automation/repository"
	"github.com/vothanh/crm-platform/internal/modules/automation/service"
)

// ProviderSet represents the dependency injection setup for the automation module.
var ProviderSet = wire.NewSet(
	repository.NewAutomationRulePostgresRepo,
	repository.NewAutomationLogPostgresRepo,
	service.NewAutomationService,
	handler.NewAutomationHandler,
)
