package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/customer/service"
	"github.com/vothanh/crm-platform/pkg/middleware"
	"github.com/vothanh/crm-platform/pkg/pagination"
	"github.com/vothanh/crm-platform/pkg/response"
)

// CustomerHandler handles HTTP requests for customer management.
type CustomerHandler struct {
	svc *service.CustomerService
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

// RegisterRoutes registers customer-related routes (tenant-scoped via middleware).
func (h *CustomerHandler) RegisterRoutes(rg *gin.RouterGroup) {
	customers := rg.Group("/customers")
	{
		customers.GET("", h.List)
		customers.GET("/:id", h.Get)
		customers.GET("/:id/timeline", h.GetTimeline)
		customers.GET("/lifecycle-counts", h.LifecycleCounts)
	}
}

// List handles GET /api/v1/customers
func (h *CustomerHandler) List(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	params := pagination.FromContext(c)
	customers, total, err := h.svc.ListCustomers(c.Request.Context(), tenantID, params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, customers, &response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: params.TotalPages(total),
	})
}

// Get handles GET /api/v1/customers/:id
func (h *CustomerHandler) Get(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	customer, err := h.svc.GetCustomer(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, customer)
}

// GetTimeline handles GET /api/v1/customers/:id/timeline
func (h *CustomerHandler) GetTimeline(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	params := pagination.FromContext(c)
	events, total, err := h.svc.GetCustomerTimeline(c.Request.Context(), tenantID, id, params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, events, &response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: params.TotalPages(total),
	})
}

// LifecycleCounts handles GET /api/v1/customers/lifecycle-counts
func (h *CustomerHandler) LifecycleCounts(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	counts, err := h.svc.GetLifecycleCounts(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, counts)
}
