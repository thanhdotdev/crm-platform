package segmentation

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/handler"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/repository"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/service"
)

// ProviderSet represents the dependency injection setup for the segmentation module.
var ProviderSet = wire.NewSet(
	repository.NewSegmentPostgresRepo,
	repository.NewCustomerSegmentPostgresRepo,
	service.NewSegmentationService,
	handler.NewSegmentationHandler,
)
