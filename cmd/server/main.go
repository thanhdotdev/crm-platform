package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/config"

	// Analytics
	analyticsHandler "github.com/vothanh/crm-platform/internal/analytics/handler"

	// Automation
	automationDomain "github.com/vothanh/crm-platform/internal/automation/domain"
	automationHandler "github.com/vothanh/crm-platform/internal/automation/handler"
	automationRepo "github.com/vothanh/crm-platform/internal/automation/repository"
	automationService "github.com/vothanh/crm-platform/internal/automation/service"

	// Campaign
	campaignDomain "github.com/vothanh/crm-platform/internal/campaign/domain"
	campaignHandler "github.com/vothanh/crm-platform/internal/campaign/handler"
	campaignRepo "github.com/vothanh/crm-platform/internal/campaign/repository"
	campaignService "github.com/vothanh/crm-platform/internal/campaign/service"

	// Customer
	customerDomain "github.com/vothanh/crm-platform/internal/customer/domain"
	customerHandler "github.com/vothanh/crm-platform/internal/customer/handler"
	customerRepo "github.com/vothanh/crm-platform/internal/customer/repository"
	customerService "github.com/vothanh/crm-platform/internal/customer/service"

	// Ingestion
	ingestionHandler "github.com/vothanh/crm-platform/internal/ingestion/handler"

	// Notification
	notificationDomain "github.com/vothanh/crm-platform/internal/notification/domain"
	notificationHandler "github.com/vothanh/crm-platform/internal/notification/handler"
	notificationProvider "github.com/vothanh/crm-platform/internal/notification/provider"
	notificationRepo "github.com/vothanh/crm-platform/internal/notification/repository"
	notificationService "github.com/vothanh/crm-platform/internal/notification/service"

	// Segmentation
	segmentationDomain "github.com/vothanh/crm-platform/internal/segmentation/domain"
	segmentationHandler "github.com/vothanh/crm-platform/internal/segmentation/handler"
	segmentationRepo "github.com/vothanh/crm-platform/internal/segmentation/repository"
	segmentationService "github.com/vothanh/crm-platform/internal/segmentation/service"

	// Tenant
	tenantDomain "github.com/vothanh/crm-platform/internal/tenant/domain"
	tenantHandler "github.com/vothanh/crm-platform/internal/tenant/handler"
	tenantRepo "github.com/vothanh/crm-platform/internal/tenant/repository"
	tenantService "github.com/vothanh/crm-platform/internal/tenant/service"

	// Trip
	tripDomain "github.com/vothanh/crm-platform/internal/trip/domain"
	tripHandler "github.com/vothanh/crm-platform/internal/trip/handler"
	tripRepo "github.com/vothanh/crm-platform/internal/trip/repository"
	tripService "github.com/vothanh/crm-platform/internal/trip/service"

	"github.com/vothanh/crm-platform/pkg/database"
	"github.com/vothanh/crm-platform/pkg/middleware"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect to database
	db, err := database.NewPostgresDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Auto-migrate all models
	err = database.AutoMigrate(db,
		// Phase 1
		&tenantDomain.Tenant{},
		&tenantDomain.APIKey{},
		&customerDomain.Customer{},
		&customerDomain.CustomerEvent{},
		&tripDomain.Trip{},
		// Phase 2
		&segmentationDomain.Segment{},
		&segmentationDomain.CustomerSegment{},
		&campaignDomain.Campaign{},
		&campaignDomain.Voucher{},
		&campaignDomain.VoucherUsage{},
		&automationDomain.AutomationRule{},
		&automationDomain.AutomationLog{},
		&notificationDomain.Notification{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 4. Initialize repositories
	tenantRepository := tenantRepo.NewTenantPostgresRepo(db)
	apiKeyRepository := tenantRepo.NewAPIKeyPostgresRepo(db)
	customerRepository := customerRepo.NewCustomerPostgresRepo(db)
	customerEventRepository := customerRepo.NewCustomerEventPostgresRepo(db)
	tripRepository := tripRepo.NewTripPostgresRepo(db)
	segmentRepository := segmentationRepo.NewSegmentPostgresRepo(db)
	customerSegmentRepository := segmentationRepo.NewCustomerSegmentPostgresRepo(db)
	campaignRepository := campaignRepo.NewCampaignPostgresRepo(db)
	voucherRepository := campaignRepo.NewVoucherPostgresRepo(db)
	voucherUsageRepository := campaignRepo.NewVoucherUsagePostgresRepo(db)
	automationRuleRepository := automationRepo.NewAutomationRulePostgresRepo(db)
	automationLogRepository := automationRepo.NewAutomationLogPostgresRepo(db)
	notificationRepository := notificationRepo.NewNotificationPostgresRepo(db)

	// 5. Initialize services
	tenantSvc := tenantService.NewTenantService(tenantRepository, apiKeyRepository)
	customerSvc := customerService.NewCustomerService(customerRepository, customerEventRepository)
	tripSvc := tripService.NewTripService(tripRepository)
	segmentationSvc := segmentationService.NewSegmentationService(segmentRepository, customerSegmentRepository)
	campaignSvc := campaignService.NewCampaignService(campaignRepository, voucherRepository, voucherUsageRepository)
	automationSvc := automationService.NewAutomationService(automationRuleRepository, automationLogRepository)

	// Mock notification providers (Phase 1)
	providers := []notificationDomain.NotificationProvider{
		notificationProvider.NewMockPushProvider(),
		notificationProvider.NewMockSMSProvider(),
		notificationProvider.NewMockEmailProvider(),
		notificationProvider.NewMockZaloProvider(),
	}
	notificationSvc := notificationService.NewNotificationService(notificationRepository, providers)

	// 6. Initialize handlers
	tenantH := tenantHandler.NewTenantHandler(tenantSvc)
	customerH := customerHandler.NewCustomerHandler(customerSvc)
	tripH := tripHandler.NewTripHandler(tripSvc)
	ingestionH := ingestionHandler.NewIngestionHandler(customerSvc, tripSvc)
	segmentationH := segmentationHandler.NewSegmentationHandler(segmentationSvc)
	campaignH := campaignHandler.NewCampaignHandler(campaignSvc)
	automationH := automationHandler.NewAutomationHandler(automationSvc)
	notificationH := notificationHandler.NewNotificationHandler(notificationSvc)
	analyticsH := analyticsHandler.NewAnalyticsHandler(db)

	// Suppress unused variable warnings
	_ = segmentationSvc
	_ = automationSvc

	// 7. Setup Gin router
	router := gin.Default()
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")

	// Admin routes (no API key needed for tenant management)
	tenantH.RegisterRoutes(v1)

	// Protected routes (API key required)
	protected := v1.Group("")
	protected.Use(middleware.APIKeyAuth(db))
	{
		// Phase 1: Core
		ingestionH.RegisterRoutes(protected)
		customerH.RegisterRoutes(protected)
		tripH.RegisterRoutes(protected)

		// Phase 2: Marketing & Intelligence
		segmentationH.RegisterRoutes(protected)
		campaignH.RegisterRoutes(protected)
		automationH.RegisterRoutes(protected)
		notificationH.RegisterRoutes(protected)
		analyticsH.RegisterRoutes(protected)
	}

	// 8. Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("[Server] Starting CRM Platform on %s", addr)
	log.Printf("[Server] Modules loaded: tenant, customer, trip, ingestion, segmentation, campaign, automation, notification, analytics")
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
