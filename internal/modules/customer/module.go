package customer

import (
	"github.com/google/wire"
	"github.com/vothanh/crm-platform/internal/modules/customer/domain"
	"github.com/vothanh/crm-platform/internal/modules/customer/handler"
	"github.com/vothanh/crm-platform/internal/modules/customer/repository"
	"github.com/vothanh/crm-platform/internal/modules/customer/service"
)

// ProviderSet represents the dependency injection setup for the customer module.
var ProviderSet = wire.NewSet(
	repository.NewCustomerPostgresRepo,
	service.NewCustomerService,
	wire.Bind(new(handler.CustomerReader), new(domain.CustomerService)),
	handler.NewCustomerHandler,
)
