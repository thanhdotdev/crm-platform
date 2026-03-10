package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/shared/middleware"
	"github.com/vothanh/crm-platform/pkg/response"
	"gorm.io/gorm"
)

// AnalyticsHandler provides dashboard metrics (per-tenant).
type AnalyticsHandler struct {
	db *gorm.DB
}

func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

func (h *AnalyticsHandler) RegisterRoutes(rg *gin.RouterGroup) {
	analytics := rg.Group("/analytics")
	{
		// Phase 2
		analytics.GET("/overview", h.Overview)
		analytics.GET("/acquisition", h.Acquisition)
		analytics.GET("/retention", h.Retention)
		// Phase 3
		analytics.GET("/cohorts", h.Cohorts)
		analytics.GET("/campaign-roi", h.CampaignROI)
		analytics.GET("/source-quality", h.SourceQuality)
		analytics.GET("/abuse-detection", h.AbuseDetection)
	}
}

// Overview returns high-level KPIs for the tenant.
func (h *AnalyticsHandler) Overview(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	var totalCustomers, totalTrips int64
	var totalRevenue float64

	h.db.Table("customers").Where("tenant_id = ?", tenantID).Count(&totalCustomers)
	h.db.Table("trips").Where("tenant_id = ? AND status = 'completed'", tenantID).Count(&totalTrips)
	h.db.Table("trips").Where("tenant_id = ? AND status = 'completed'", tenantID).
		Select("COALESCE(SUM(amount), 0)").Row().Scan(&totalRevenue)

	// Lifecycle breakdown
	type lifecycleCount struct {
		Stage string `json:"stage"`
		Count int64  `json:"count"`
	}
	var lifecycleCounts []lifecycleCount
	h.db.Table("customers").
		Select("lifecycle_stage as stage, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("lifecycle_stage").
		Find(&lifecycleCounts)

	// Tier breakdown
	type tierCount struct {
		Tier  string `json:"tier"`
		Count int64  `json:"count"`
	}
	var tierCounts []tierCount
	h.db.Table("customers").
		Select("tier, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("tier").
		Find(&tierCounts)

	response.OK(c, gin.H{
		"total_customers":      totalCustomers,
		"total_trips":          totalTrips,
		"total_revenue":        totalRevenue,
		"avg_revenue_per_trip": safeDiv(totalRevenue, float64(totalTrips)),
		"lifecycle_breakdown":  lifecycleCounts,
		"tier_breakdown":       tierCounts,
	})
}

// Acquisition returns source-based analytics.
func (h *AnalyticsHandler) Acquisition(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	type sourceMetrics struct {
		Source   string  `json:"source"`
		Count    int64   `json:"count"`
		AvgTrips float64 `json:"avg_trips"`
		AvgSpent float64 `json:"avg_spent"`
		AvgScore float64 `json:"avg_lead_score"`
	}
	var metrics []sourceMetrics
	h.db.Table("customers").
		Select("source, COUNT(*) as count, AVG(total_trips) as avg_trips, AVG(total_spent) as avg_spent, AVG(lead_score) as avg_score").
		Where("tenant_id = ? AND source != ''", tenantID).
		Group("source").
		Order("count DESC").
		Find(&metrics)

	response.OK(c, metrics)
}

// Retention returns retention metrics.
func (h *AnalyticsHandler) Retention(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	type retentionMetric struct {
		TripBucket string `json:"trip_bucket"`
		Count      int64  `json:"count"`
	}
	var metrics []retentionMetric
	h.db.Table("customers").
		Select(`CASE
			WHEN total_trips = 0 THEN '0_trips'
			WHEN total_trips = 1 THEN '1_trip'
			WHEN total_trips = 2 THEN '2_trips'
			WHEN total_trips >= 3 AND total_trips < 5 THEN '3-4_trips'
			WHEN total_trips >= 5 THEN '5+_trips'
		END as trip_bucket, COUNT(*) as count`).
		Where("tenant_id = ?", tenantID).
		Group("trip_bucket").
		Order("trip_bucket").
		Find(&metrics)

	// Conversion rates
	var registered, oneTrip, twoTrips, threeTrips int64
	h.db.Table("customers").Where("tenant_id = ?", tenantID).Count(&registered)
	h.db.Table("customers").Where("tenant_id = ? AND total_trips >= 1", tenantID).Count(&oneTrip)
	h.db.Table("customers").Where("tenant_id = ? AND total_trips >= 2", tenantID).Count(&twoTrips)
	h.db.Table("customers").Where("tenant_id = ? AND total_trips >= 3", tenantID).Count(&threeTrips)

	response.OK(c, gin.H{
		"trip_distribution": metrics,
		"conversion_funnel": gin.H{
			"registered_to_1st_trip": safePercent(oneTrip, registered),
			"1st_to_2nd_trip":        safePercent(twoTrips, oneTrip),
			"2nd_to_3rd_trip":        safePercent(threeTrips, twoTrips),
		},
	})
}

func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func safePercent(n, d int64) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d) * 100
}

// ========== PHASE 3 ==========

// Cohorts returns weekly registration cohorts with retention rates.
// Shows: users registered in week X → how many completed 1, 2, 3+ trips.
func (h *AnalyticsHandler) Cohorts(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	// Default: last 12 weeks
	weeks := 12
	if w := c.Query("weeks"); w != "" {
		if _, err := time.Parse("2006", w); err == nil {
			weeks = 52
		} // just use default
	}

	type cohortRow struct {
		CohortWeek     string  `json:"cohort_week"`
		Registered     int64   `json:"registered"`
		Activated      int64   `json:"activated"` // ≥1 trip
		Returning      int64   `json:"returning"` // ≥2 trips
		Loyal          int64   `json:"loyal"`     // ≥3 trips
		ActivationRate float64 `json:"activation_rate"`
		RetentionRate  float64 `json:"retention_rate"`
	}

	var rows []cohortRow
	sinceDate := time.Now().AddDate(0, 0, -7*weeks)

	h.db.Raw(`
		SELECT
			TO_CHAR(DATE_TRUNC('week', created_at), 'YYYY-WW') as cohort_week,
			COUNT(*) as registered,
			COUNT(*) FILTER (WHERE total_trips >= 1) as activated,
			COUNT(*) FILTER (WHERE total_trips >= 2) as returning,
			COUNT(*) FILTER (WHERE total_trips >= 3) as loyal
		FROM customers
		WHERE tenant_id = ? AND created_at >= ?
		GROUP BY cohort_week
		ORDER BY cohort_week
	`, tenantID, sinceDate).Scan(&rows)

	// Calculate rates
	for i := range rows {
		rows[i].ActivationRate = safePercent(rows[i].Activated, rows[i].Registered)
		rows[i].RetentionRate = safePercent(rows[i].Returning, rows[i].Activated)
	}

	response.OK(c, rows)
}

// CampaignROI returns ROI metrics for campaigns.
func (h *AnalyticsHandler) CampaignROI(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	// Optional campaign_id filter
	var campaignFilter string
	var args []interface{}
	args = append(args, tenantID)
	if cid := c.Query("campaign_id"); cid != "" {
		if id, err := strconv.ParseUint(cid, 10, 64); err == nil {
			campaignFilter = " AND v.campaign_id = ?"
			args = append(args, id)
		}
	}

	type roiRow struct {
		CampaignID     string  `json:"campaign_id"`
		CampaignName   string  `json:"campaign_name"`
		TotalVouchers  int64   `json:"total_vouchers"`
		UsedVouchers   int64   `json:"used_vouchers"`
		RedemptionRate float64 `json:"redemption_rate"`
		TotalDiscount  float64 `json:"total_discount_given"`
	}

	var rows []roiRow
	h.db.Raw(`
		SELECT
			v.campaign_id::text as campaign_id,
			c.name as campaign_name,
			COUNT(v.id) as total_vouchers,
			SUM(CASE WHEN v.used_count > 0 THEN 1 ELSE 0 END) as used_vouchers,
			SUM(CASE WHEN v.used_count > 0 THEN v.discount_value ELSE 0 END) as total_discount
		FROM vouchers v
		JOIN campaigns c ON c.id = v.campaign_id
		WHERE v.tenant_id = ?`+campaignFilter+`
		GROUP BY v.campaign_id, c.name
		ORDER BY total_vouchers DESC
	`, args...).Scan(&rows)

	for i := range rows {
		rows[i].RedemptionRate = safePercent(rows[i].UsedVouchers, rows[i].TotalVouchers)
	}

	response.OK(c, rows)
}

// SourceQuality returns acquisition source quality metrics.
func (h *AnalyticsHandler) SourceQuality(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	type sourceRow struct {
		Source         string  `json:"source"`
		Customers      int64   `json:"customers"`
		Activated      int64   `json:"activated"`
		ConversionRate float64 `json:"conversion_rate"`
		AvgTrips       float64 `json:"avg_trips"`
		AvgSpent       float64 `json:"avg_spent"`
		TotalRevenue   float64 `json:"total_revenue"`
		LTV            float64 `json:"ltv"`
		LuxuryCount    int64   `json:"luxury_count"`
		LuxuryRate     float64 `json:"luxury_rate"`
		AvgLeadScore   float64 `json:"avg_lead_score"`
		CancelledTrips int64   `json:"cancelled_trips"`
		CancelRate     float64 `json:"cancel_rate"`
	}

	var rows []sourceRow
	h.db.Raw(`
		SELECT
			c.source,
			COUNT(*) as customers,
			COUNT(*) FILTER (WHERE c.total_trips >= 1) as activated,
			AVG(c.total_trips) as avg_trips,
			AVG(c.total_spent) as avg_spent,
			SUM(c.total_spent) as total_revenue,
			CASE WHEN COUNT(*) > 0 THEN SUM(c.total_spent) / COUNT(*) ELSE 0 END as ltv,
			COUNT(*) FILTER (WHERE c.tier = 'luxury') as luxury_count,
			AVG(c.lead_score) as avg_lead_score
		FROM customers c
		WHERE c.tenant_id = ? AND c.source != ''
		GROUP BY c.source
		ORDER BY total_revenue DESC
	`, tenantID).Scan(&rows)

	// Calculate rates and get cancel data
	for i := range rows {
		rows[i].ConversionRate = safePercent(rows[i].Activated, rows[i].Customers)
		rows[i].LuxuryRate = safePercent(rows[i].LuxuryCount, rows[i].Customers)

		// Get cancel stats per source
		var completedTrips, cancelledTrips int64
		h.db.Raw(`
			SELECT
				COUNT(*) FILTER (WHERE t.status = 'completed') as completed,
				COUNT(*) FILTER (WHERE t.status = 'cancelled') as cancelled
			FROM trips t
			JOIN customers c ON c.id = t.customer_id AND c.tenant_id = t.tenant_id
			WHERE t.tenant_id = ? AND c.source = ?
		`, tenantID, rows[i].Source).Row().Scan(&completedTrips, &cancelledTrips)
		rows[i].CancelledTrips = cancelledTrips
		rows[i].CancelRate = safePercent(cancelledTrips, completedTrips+cancelledTrips)
	}

	response.OK(c, rows)
}

// AbuseDetection returns suspicious customer patterns.
func (h *AnalyticsHandler) AbuseDetection(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		response.Unauthorized(c, "tenant not resolved")
		return
	}

	type suspiciousCustomer struct {
		CustomerID   string `json:"customer_id"`
		ExternalID   string `json:"external_id"`
		FullName     string `json:"full_name"`
		Reason       string `json:"reason"`
		RiskScore    int    `json:"risk_score"`
		TotalTrips   int    `json:"total_trips"`
		CancelCount  int    `json:"cancel_count"`
		VoucherCount int    `json:"voucher_count"`
	}

	var results []suspiciousCustomer

	// 1. High cancel rate (≥50% of trips cancelled, min 3 total trips+cancels)
	type cancelAbuser struct {
		CustomerID string `json:"customer_id"`
		ExternalID string `json:"external_id"`
		FullName   string `json:"full_name"`
		Completed  int    `json:"completed"`
		Cancelled  int    `json:"cancelled"`
	}
	var cancelAbusers []cancelAbuser
	h.db.Raw(`
		SELECT
			c.id::text as customer_id,
			c.external_id,
			c.full_name,
			COUNT(*) FILTER (WHERE t.status = 'completed') as completed,
			COUNT(*) FILTER (WHERE t.status = 'cancelled') as cancelled
		FROM customers c
		JOIN trips t ON t.customer_id = c.id AND t.tenant_id = c.tenant_id
		WHERE c.tenant_id = ?
		GROUP BY c.id, c.external_id, c.full_name
		HAVING COUNT(*) FILTER (WHERE t.status = 'cancelled') >= 2
			AND COUNT(*) FILTER (WHERE t.status = 'cancelled')::float / GREATEST(COUNT(*), 1) >= 0.5
		ORDER BY cancelled DESC
	`, tenantID).Scan(&cancelAbusers)

	for _, ca := range cancelAbusers {
		score := 30
		if ca.Cancelled >= 5 {
			score = 70
		} else if ca.Cancelled >= 3 {
			score = 50
		}
		results = append(results, suspiciousCustomer{
			CustomerID:  ca.CustomerID,
			ExternalID:  ca.ExternalID,
			FullName:    ca.FullName,
			Reason:      "high_cancel_rate",
			RiskScore:   score,
			TotalTrips:  ca.Completed,
			CancelCount: ca.Cancelled,
		})
	}

	// 2. Rapid voucher usage (used ≥3 vouchers in 7 days)
	type voucherAbuser struct {
		CustomerID   string `json:"customer_id"`
		ExternalID   string `json:"external_id"`
		FullName     string `json:"full_name"`
		VoucherCount int    `json:"voucher_count"`
	}
	var voucherAbusers []voucherAbuser
	h.db.Raw(`
		SELECT
			c.id::text as customer_id,
			c.external_id,
			c.full_name,
			COUNT(vu.id) as voucher_count
		FROM customers c
		JOIN voucher_usages vu ON vu.customer_id = c.id AND vu.tenant_id = c.tenant_id
		WHERE c.tenant_id = ? AND vu.used_at >= NOW() - INTERVAL '7 days'
		GROUP BY c.id, c.external_id, c.full_name
		HAVING COUNT(vu.id) >= 3
		ORDER BY voucher_count DESC
	`, tenantID).Scan(&voucherAbusers)

	for _, va := range voucherAbusers {
		score := 60
		if va.VoucherCount >= 5 {
			score = 90
		}
		results = append(results, suspiciousCustomer{
			CustomerID:   va.CustomerID,
			ExternalID:   va.ExternalID,
			FullName:     va.FullName,
			Reason:       "rapid_voucher_usage",
			RiskScore:    score,
			VoucherCount: va.VoucherCount,
		})
	}

	// 3. Summary stats
	var totalCustomers, flaggedCount int64
	h.db.Table("customers").Where("tenant_id = ?", tenantID).Count(&totalCustomers)
	flaggedCount = int64(len(results))

	response.OK(c, gin.H{
		"total_customers":  totalCustomers,
		"flagged_count":    flaggedCount,
		"flagged_rate":     safePercent(flaggedCount, totalCustomers),
		"suspicious_users": results,
	})
}
