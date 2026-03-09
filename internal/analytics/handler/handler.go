package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/pkg/middleware"
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
		analytics.GET("/overview", h.Overview)
		analytics.GET("/acquisition", h.Acquisition)
		analytics.GET("/retention", h.Retention)
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
