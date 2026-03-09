//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/infra/config"
	"github.com/vothanh/crm-platform/internal/infra/database"

	customerDomain "github.com/vothanh/crm-platform/internal/modules/customer/domain"
	customerHandler "github.com/vothanh/crm-platform/internal/modules/customer/handler"
	ingestionDomain "github.com/vothanh/crm-platform/internal/modules/ingestion/domain"
	ingestionHandler "github.com/vothanh/crm-platform/internal/modules/ingestion/handler"
	tripService "github.com/vothanh/crm-platform/internal/modules/trip/service"

	"github.com/vothanh/crm-platform/internal/modules/analytics"
	"github.com/vothanh/crm-platform/internal/modules/automation"
	"github.com/vothanh/crm-platform/internal/modules/campaign"
	"github.com/vothanh/crm-platform/internal/modules/customer"
	"github.com/vothanh/crm-platform/internal/modules/ingestion"
	"github.com/vothanh/crm-platform/internal/modules/notification"
	"github.com/vothanh/crm-platform/internal/modules/segmentation"
	"github.com/vothanh/crm-platform/internal/modules/tenant"
	"github.com/vothanh/crm-platform/internal/modules/trip"
)

// CrossModuleBindings provides interfaces that span across different modules.
var CrossModuleBindings = wire.NewSet(
	provideCustomerActuator,
	provideTripActuator,
	provideEventReader,
)

func provideCustomerActuator(svc customerDomain.CustomerService) ingestionHandler.CustomerActuator {
	return svc
}

func provideTripActuator(svc *tripService.TripService) ingestionHandler.TripActuator {
	return svc
}

func provideEventReader(repo ingestionDomain.EventRepository) customerHandler.EventReader {
	return repo
}

// provideDBConfig extracts the DSN string from the config.
func provideDBConfig(cfg *config.Config) string {
	return cfg.Database.DSN()
}

// InitializeApp sets up the entire application using Google Wire.
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		// Infrastructure
		provideDBConfig,
		database.NewPostgresDB,

		// Modules
		tenant.ProviderSet,
		customer.ProviderSet,
		trip.ProviderSet,
		ingestion.ProviderSet,
		segmentation.ProviderSet,
		campaign.ProviderSet,
		automation.ProviderSet,
		notification.ProviderSet,
		analytics.ProviderSet,

		// Bindings across modules
		CrossModuleBindings,

		// Core App
		NewAppWithDependencies,
	)
	return &App{}, nil
}
