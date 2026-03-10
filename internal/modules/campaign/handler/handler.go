package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/modules/campaign/domain"
	"github.com/vothanh/crm-platform/internal/modules/campaign/service"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/pagination"
	"github.com/vothanh/crm-platform/pkg/response"
)

type CampaignHandler struct {
	svc *service.CampaignService
}

func NewCampaignHandler(svc *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{svc: svc}
}

func (h *CampaignHandler) RegisterRoutes(rg *gin.RouterGroup) {
	campaigns := rg.Group("/campaigns")
	{
		campaigns.POST("", h.Create)
		campaigns.GET("", h.List)
		campaigns.GET("/:id", h.Get)
		campaigns.GET("/:id/vouchers", h.ListVouchers)
		campaigns.POST("/:id/vouchers/generate", h.GenerateVoucher)
		campaigns.POST("/vouchers/redeem", h.RedeemVoucher)
	}
}

type createCampaignRequest struct {
	Name            string  `json:"name" binding:"required"`
	Type            string  `json:"type" binding:"required"`
	TargetSegmentID *uint64 `json:"target_segment_id"`
	Budget          float64 `json:"budget"`
	StartDate       string  `json:"start_date" binding:"required"`
	EndDate         string  `json:"end_date" binding:"required"`
	Description     string  `json:"description"`
}

func (h *CampaignHandler) Create(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var req createCampaignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	startDate, _ := parseTime(req.StartDate)
	endDate, _ := parseTime(req.EndDate)

	campaign := &domain.Campaign{
		TenantID:        tenantID,
		Name:            req.Name,
		Type:            req.Type,
		Status:          "draft",
		TargetSegmentID: req.TargetSegmentID,
		Budget:          req.Budget,
		StartDate:       startDate,
		EndDate:         endDate,
		Description:     req.Description,
	}

	result, err := h.svc.CreateCampaign(c.Request.Context(), campaign)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, result)
}

func (h *CampaignHandler) List(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	params := pagination.FromContext(c)
	campaigns, total, err := h.svc.ListCampaigns(c.Request.Context(), tenantID, params.Offset(), params.PerPage)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMeta(c, http.StatusOK, campaigns, &response.Meta{
		Page: params.Page, PerPage: params.PerPage, Total: total, TotalPages: params.TotalPages(total),
	})
}

func (h *CampaignHandler) Get(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid campaign ID")
		return
	}

	campaign, err := h.svc.GetCampaign(c.Request.Context(), tenantID, id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.OK(c, campaign)
}

type generateVoucherRequest struct {
	CustomerID uint64 `json:"customer_id" binding:"required"`
	TripNumber int    `json:"trip_number" binding:"required,min=1,max=3"`
}

func (h *CampaignHandler) GenerateVoucher(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	campaignID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid campaign ID")
		return
	}

	var req generateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	voucher, err := h.svc.GeneratePolicy50_30_20Voucher(c.Request.Context(), tenantID, campaignID, req.CustomerID, req.TripNumber)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Created(c, voucher)
}

func (h *CampaignHandler) ListVouchers(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	campaignID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid campaign ID")
		return
	}

	vouchers, err := h.svc.ListVouchers(c.Request.Context(), tenantID, campaignID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, vouchers)
}

type redeemVoucherRequest struct {
	Code       string `json:"code" binding:"required"`
	CustomerID uint64 `json:"customer_id" binding:"required"`
	TripID     uint64 `json:"trip_id" binding:"required"`
}

func (h *CampaignHandler) RedeemVoucher(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var req redeemVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	voucher, err := h.svc.RedeemVoucher(c.Request.Context(), tenantID, req.Code, req.CustomerID, req.TripID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "REDEEM_FAILED", err.Error())
		return
	}

	response.OK(c, voucher)
}

func parseTime(s string) (time.Time, error) {
	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", s)
}
