package ingestion

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/ingestion/handler"
	"github.com/vothanh/crm-platform/internal/modules/ingestion/repository"
	"github.com/vothanh/crm-platform/internal/modules/ingestion/service"
)

// ProviderSet represents the dependency injection setup for the ingestion module.
var ProviderSet = wire.NewSet(
	repository.NewEventPostgresRepo,
	service.NewIngestionService,
	handler.NewIngestionHandler,
)
