package analytics

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/analytics/handler"
)

// ProviderSet represents the dependency injection setup for the analytics module.
var ProviderSet = wire.NewSet(
	handler.NewAnalyticsHandler,
)
