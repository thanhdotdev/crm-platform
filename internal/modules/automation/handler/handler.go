package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/modules/automation/domain"
	"github.com/vothanh/crm-platform/internal/modules/automation/service"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/pagination"
	"github.com/vothanh/crm-platform/pkg/response"
	"gorm.io/datatypes"
)

type AutomationHandler struct {
	svc *service.AutomationService
}

func NewAutomationHandler(svc *service.AutomationService) *AutomationHandler {
	return &AutomationHandler{svc: svc}
}

func (h *AutomationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rules := rg.Group("/automation/rules")
	{
		rules.POST("", h.CreateRule)
		rules.GET("", h.ListRules)
		rules.GET("/:id", h.GetRule)
		rules.PUT("/:id", h.UpdateRule)
		rules.GET("/:id/logs", h.ListLogs)
	}
}

type createRuleRequest struct {
	Name            string         `json:"name" binding:"required"`
	TriggerType     string         `json:"trigger_type" binding:"required"`
	TriggerConfig   datatypes.JSON `json:"trigger_config"`
	ActionType      string         `json:"action_type" binding:"required"`
	ActionConfig    datatypes.JSON `json:"action_config"`
	TargetSegmentID *uint64        `json:"target_segment_id"`
	ExecutionMode   string         `json:"execution_mode"` // automatic, manual
}

func (h *AutomationHandler) CreateRule(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	mode := req.ExecutionMode
	if mode == "" {
		mode = "automatic"
	}

	rule := &domain.AutomationRule{
		TenantID:        tenantID,
		Name:            req.Name,
		TriggerType:     req.TriggerType,
		TriggerConfig:   req.TriggerConfig,
		ActionType:      req.ActionType,
		ActionConfig:    req.ActionConfig,
		TargetSegmentID: req.TargetSegmentID,
		IsActive:        true,
		ExecutionMode:   mode,
	}

	result, err := h.svc.CreateRule(c.Request.Context(), rule)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, result)
}

func (h *AutomationHandler) ListRules(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	rules, err := h.svc.ListRules(c.Request.Context(), tenantID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, rules)
}

func (h *AutomationHandler) GetRule(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid rule ID")
		return
	}

	rule, err := h.svc.GetRule(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, rule)
}

func (h *AutomationHandler) UpdateRule(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid rule ID")
		return
	}

	existing, err := h.svc.GetRule(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	existing.Name = req.Name
	existing.TriggerType = req.TriggerType
	existing.TriggerConfig = req.TriggerConfig
	existing.ActionType = req.ActionType
	existing.ActionConfig = req.ActionConfig
	existing.TargetSegmentID = req.TargetSegmentID
	if req.ExecutionMode != "" {
		existing.ExecutionMode = req.ExecutionMode
	}

	result, err := h.svc.UpdateRule(c.Request.Context(), existing)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *AutomationHandler) ListLogs(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid rule ID")
		return
	}

	params := pagination.FromContext(c)
	logs, total, err := h.svc.ListLogs(c.Request.Context(), tenantID, ruleID, params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, logs, &response.Meta{
		Page: params.Page, PerPage: params.PerPage, Total: total, TotalPages: params.TotalPages(total),
	})
}
