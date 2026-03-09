package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/notification/domain"
	"github.com/vothanh/crm-platform/internal/modules/notification/service"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/pagination"
	"github.com/vothanh/crm-platform/pkg/response"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	notifications := rg.Group("/notifications")
	{
		notifications.POST("", h.Send)
		notifications.GET("/customer/:customer_id", h.ListByCustomer)
		notifications.GET("/channels", h.ListChannels)
	}
}

type sendRequest struct {
	CustomerID uuid.UUID `json:"customer_id" binding:"required"`
	Channel    string    `json:"channel" binding:"required"`
	Title      string    `json:"title" binding:"required"`
	Content    string    `json:"content" binding:"required"`
}

func (h *NotificationHandler) Send(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var req sendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	n := &domain.Notification{
		TenantID:   tenantID,
		CustomerID: req.CustomerID,
		Channel:    req.Channel,
		Title:      req.Title,
		Content:    req.Content,
	}

	result, err := h.svc.Send(c.Request.Context(), n)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, result)
}

func (h *NotificationHandler) ListByCustomer(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	customerID, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		response.BadRequest(c, "invalid customer ID")
		return
	}

	params := pagination.FromContext(c)
	notifications, total, err := h.svc.ListByCustomer(c.Request.Context(), tenantID, customerID, params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, notifications, &response.Meta{
		Page: params.Page, PerPage: params.PerPage, Total: total, TotalPages: params.TotalPages(total),
	})
}

func (h *NotificationHandler) ListChannels(c *gin.Context) {
	response.OK(c, h.svc.AvailableChannels())
}
