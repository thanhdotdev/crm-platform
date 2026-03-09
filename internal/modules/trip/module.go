package trip

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/trip/handler"
	"github.com/vothanh/crm-platform/internal/modules/trip/repository"
	"github.com/vothanh/crm-platform/internal/modules/trip/service"
)

// ProviderSet represents the dependency injection setup for the trip module.
var ProviderSet = wire.NewSet(
	repository.NewTripPostgresRepo,
	service.NewTripService,
	handler.NewTripHandler,
)
