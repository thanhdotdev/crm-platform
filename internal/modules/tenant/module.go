package tenant

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/tenant/handler"
	"github.com/vothanh/crm-platform/internal/modules/tenant/repository"
	"github.com/vothanh/crm-platform/internal/modules/tenant/service"
)

// ProviderSet represents the dependency injection setup for the tenant module.
var ProviderSet = wire.NewSet(
	repository.NewTenantPostgresRepo,
	repository.NewAPIKeyPostgresRepo,
	service.NewTenantService,
	handler.NewTenantHandler,
)
