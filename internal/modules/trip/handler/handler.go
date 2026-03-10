package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/modules/trip/service"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/pagination"
	"github.com/vothanh/crm-platform/pkg/response"
)

// TripHandler handles HTTP requests for trip management.
type TripHandler struct {
	svc *service.TripService
}

// NewTripHandler creates a new TripHandler.
func NewTripHandler(svc *service.TripService) *TripHandler {
	return &TripHandler{svc: svc}
}

// RegisterRoutes registers trip-related routes.
func (h *TripHandler) RegisterRoutes(rg *gin.RouterGroup) {
	trips := rg.Group("/trips")
	{
		trips.GET("/:id", h.Get)
		trips.GET("/customer/:customer_id", h.ListByCustomer)
	}
}

// Get handles GET /api/v1/trips/:id
func (h *TripHandler) Get(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid trip ID")
		return
	}

	trip, err := h.svc.GetTrip(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, trip)
}

// ListByCustomer handles GET /api/v1/trips/customer/:customer_id
func (h *TripHandler) ListByCustomer(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	params := pagination.FromContext(c)
	trips, total, err := h.svc.ListByCustomer(c.Request.Context(), tenantID, customerID, params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, trips, &response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: params.TotalPages(total),
	})
}
