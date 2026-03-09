package campaign

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/campaign/handler"
	"github.com/vothanh/crm-platform/internal/modules/campaign/repository"
	"github.com/vothanh/crm-platform/internal/modules/campaign/service"
)

// ProviderSet represents the dependency injection setup for the campaign module.
var ProviderSet = wire.NewSet(
	repository.NewCampaignPostgresRepo,
	repository.NewVoucherPostgresRepo,
	repository.NewVoucherUsagePostgresRepo,
	service.NewCampaignService,
	handler.NewCampaignHandler,
)
