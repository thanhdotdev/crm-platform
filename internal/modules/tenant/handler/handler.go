package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/tenant/service"
	"github.com/vothanh/crm-platform/pkg/pagination"
	"github.com/vothanh/crm-platform/pkg/response"
)

// TenantHandler handles HTTP requests for tenant management.
type TenantHandler struct {
	svc *service.TenantService
}

// NewTenantHandler creates a new TenantHandler.
func NewTenantHandler(svc *service.TenantService) *TenantHandler {
	return &TenantHandler{svc: svc}
}

// RegisterRoutes registers tenant-related routes.
func (h *TenantHandler) RegisterRoutes(rg *gin.RouterGroup) {
	tenants := rg.Group("/tenants")
	{
		tenants.POST("", h.Create)
		tenants.GET("", h.List)
		tenants.GET("/:id", h.Get)
		tenants.PATCH("/:id", h.Update)
	}

	apiKeys := rg.Group("/tenants/:id/api-keys")
	{
		apiKeys.POST("", h.CreateAPIKey)
		apiKeys.GET("", h.ListAPIKeys)
	}
}

// createTenantRequest represents the create tenant request body.
type createTenantRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// Create handles POST /api/v1/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tenant, err := h.svc.CreateTenant(c.Request.Context(), req.Name, req.Slug)
	if err != nil {
		response.Error(c, http.StatusConflict, "CONFLICT", err.Error())
		return
	}

	response.Created(c, tenant)
}

// Get handles GET /api/v1/tenants/:id
func (h *TenantHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tenant ID")
		return
	}

	tenant, err := h.svc.GetTenant(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, tenant)
}

// List handles GET /api/v1/tenants
func (h *TenantHandler) List(c *gin.Context) {
	params := pagination.FromContext(c)

	tenants, total, err := h.svc.ListTenants(c.Request.Context(), params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, tenants, &response.Meta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: params.TotalPages(total),
	})
}

// updateTenantRequest represents the update tenant request body.
type updateTenantRequest struct {
	Name     string `json:"name"`
	IsActive *bool  `json:"is_active"`
}

// Update handles PATCH /api/v1/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tenant ID")
		return
	}

	var req updateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tenant, err := h.svc.UpdateTenant(c.Request.Context(), id, req.Name, req.IsActive)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, tenant)
}

// createAPIKeyRequest represents the create API key request body.
type createAPIKeyRequest struct {
	KeyType string `json:"key_type" binding:"required,oneof=server client"`
	Name    string `json:"name" binding:"required"`
}

// CreateAPIKey handles POST /api/v1/tenants/:id/api-keys
func (h *TenantHandler) CreateAPIKey(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tenant ID")
		return
	}

	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	key, err := h.svc.CreateAPIKey(c.Request.Context(), tenantID, req.KeyType, req.Name)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, key)
}

// ListAPIKeys handles GET /api/v1/tenants/:id/api-keys
func (h *TenantHandler) ListAPIKeys(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tenant ID")
		return
	}

	keys, err := h.svc.ListAPIKeys(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, keys)
}
