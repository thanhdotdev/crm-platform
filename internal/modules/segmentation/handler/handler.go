package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/domain"
	"github.com/vothanh/crm-platform/internal/modules/segmentation/service"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/response"
	"gorm.io/datatypes"
)

// SegmentationHandler handles HTTP requests for segmentation.
type SegmentationHandler struct {
	svc *service.SegmentationService
}

func NewSegmentationHandler(svc *service.SegmentationService) *SegmentationHandler {
	return &SegmentationHandler{svc: svc}
}

func (h *SegmentationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	segments := rg.Group("/segments")
	{
		segments.POST("", h.Create)
		segments.GET("", h.List)
		segments.GET("/:id", h.Get)
		segments.PUT("/:id", h.Update)
	}
}

type createSegmentRequest struct {
	Name        string         `json:"name" binding:"required"`
	Type        string         `json:"type" binding:"required"`
	Rules       datatypes.JSON `json:"rules"`
	RiskLevel   string         `json:"risk_level"`
	Description string         `json:"description"`
}

func (h *SegmentationHandler) Create(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var req createSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	segment := &domain.Segment{
		TenantID:    tenantID,
		Name:        req.Name,
		Type:        domain.SegmentType(req.Type),
		Rules:       req.Rules,
		RiskLevel:   req.RiskLevel,
		Description: req.Description,
		IsActive:    true,
	}

	result, err := h.svc.CreateSegment(c.Request.Context(), segment)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, result)
}

func (h *SegmentationHandler) List(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	segments, err := h.svc.ListSegments(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, segments)
}

func (h *SegmentationHandler) Get(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid segment ID")
		return
	}

	segment, err := h.svc.GetSegment(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	// Include customer count
	count, _ := h.svc.GetSegmentCount(c.Request.Context(), tenantID, id)

	response.OK(c, gin.H{
		"segment":        segment,
		"customer_count": count,
	})
}

func (h *SegmentationHandler) Update(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid segment ID")
		return
	}

	existing, err := h.svc.GetSegment(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var req createSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	existing.Name = req.Name
	existing.Type = domain.SegmentType(req.Type)
	existing.Rules = req.Rules
	existing.RiskLevel = req.RiskLevel
	existing.Description = req.Description

	result, err := h.svc.UpdateSegment(c.Request.Context(), existing)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}
