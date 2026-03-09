package notification

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
	"github.com/vothanh/crm-platform/internal/modules/notification/handler"
	"github.com/vothanh/crm-platform/internal/modules/notification/provider"
	"github.com/vothanh/crm-platform/internal/modules/notification/repository"
	"github.com/vothanh/crm-platform/internal/modules/notification/service"
)

// ProviderSet represents the dependency injection setup for the notification module.
var ProviderSet = wire.NewSet(
	repository.NewNotificationPostgresRepo,
	ProvideNotificationProviders,
	service.NewNotificationService,
	handler.NewNotificationHandler,
)

// ProvideNotificationProviders is an interim provider to group the mock providers.
func ProvideNotificationProviders() []domain.NotificationProvider {
	return []domain.NotificationProvider{
		provider.NewMockPushProvider(),
		provider.NewMockSMSProvider(),
		provider.NewMockEmailProvider(),
		provider.NewMockZaloProvider(),
	}
}
