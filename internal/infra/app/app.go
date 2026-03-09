package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/infra/config"
	"github.com/vothanh/crm-platform/internal/infra/database"
	"gorm.io/gorm"

	// Analytics
	analyticsHandler "github.com/vothanh/crm-platform/internal/modules/analytics/handler"

	// Automation
	automationDomain "github.com/vothanh/crm-platform/internal/modules/automation/domain"
	automationHandler "github.com/vothanh/crm-platform/internal/modules/automation/handler"
	automationRepo "github.com/vothanh/crm-platform/internal/modules/automation/repository"
	automationService "github.com/vothanh/crm-platform/internal/modules/automation/service"

	// Campaign
	campaignDomain "github.com/vothanh/crm-platform/internal/modules/campaign/domain"
	campaignHandler "github.com/vothanh/crm-platform/internal/modules/campaign/handler"
	campaignRepo "github.com/vothanh/crm-platform/internal/modules/campaign/repository"
	campaignService "github.com/vothanh/crm-platform/internal/modules/campaign/service"

	// Customer
	customerDomain "github.com/vothanh/crm-platform/internal/modules/customer/domain"
	customerHandler "github.com/vothanh/crm-platform/internal/modules/customer/handler"
	customerRepo "github.com/vothanh/crm-platform/internal/modules/customer/repository"
	customerService "github.com/vothanh/crm-platform/internal/modules/customer/service"

	// Ingestion
	ingestionHandler "github.com/vothanh/crm-platform/internal/modules/ingestion/handler"

	// Notification
	notificationDomain "github.com/vothanh/crm-platform/internal/modules/notification/domain"
	notificationHandler "github.com/vothanh/crm-platform/internal/modules/notification/handler"
	notificationProvider "github.com/vothanh/crm-platform/internal/modules/notification/provider"
	notificationRepo "github.com/vothanh/crm-platform/internal/modules/notification/repository"
	notificationService "github.com/vothanh/crm-platform/internal/modules/notification/service"

	// Segmentation
	segmentationDomain "github.com/vothanh/crm-platform/internal/modules/segmentation/domain"
	segmentationHandler "github.com/vothanh/crm-platform/internal/modules/segmentation/handler"
	segmentationRepo "github.com/vothanh/crm-platform/internal/modules/segmentation/repository"
	segmentationService "github.com/vothanh/crm-platform/internal/modules/segmentation/service"

	// Tenant
	tenantDomain "github.com/vothanh/crm-platform/internal/modules/tenant/domain"
	tenantHandler "github.com/vothanh/crm-platform/internal/modules/tenant/handler"
	tenantRepo "github.com/vothanh/crm-platform/internal/modules/tenant/repository"
	tenantService "github.com/vothanh/crm-platform/internal/modules/tenant/service"

	// Trip
	tripDomain "github.com/vothanh/crm-platform/internal/modules/trip/domain"
	tripHandler "github.com/vothanh/crm-platform/internal/modules/trip/handler"
	tripRepo "github.com/vothanh/crm-platform/internal/modules/trip/repository"
	tripService "github.com/vothanh/crm-platform/internal/modules/trip/service"

	"github.com/vothanh/crm-platform/internal/shared/middleware"
)

// App manages the lifecycle, routing, and dependency injection of the CRM.
type App struct {
	cfg    *config.Config
	db     *gorm.DB
	router *gin.Engine
	server *http.Server
}

// NewApp initializes dependencies, connects to the database, and sets up routing.
func NewApp(cfg *config.Config) (*App, error) {
	// Connect to database
	db, err := database.NewPostgresDB(cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	app := &App{
		cfg: cfg,
		db:  db,
	}

	if err := app.migrate(); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	app.setupRouter()
	return app, nil
}

// Run starts the HTTP server and handles graceful shutdown.
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.cfg.Port)
	a.server = &http.Server{
		Addr:    addr,
		Handler: a.router,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("[Server] Starting CRM Platform on %s", addr)
		if err := a.server.ListenAndServe(); string(err.Error()) != "http: Server closed" {
			serverErrors <- err
		}
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block main waiting for shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("[Server] Shutdown signal received: %v. Initiating graceful shutdown...", sig)

		// Ask listener to shut down in 10 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := a.server.Shutdown(ctx); err != nil {
			a.server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func (a *App) migrate() error {
	return database.AutoMigrate(a.db,
		// Phase 1
		&tenantDomain.Tenant{},
		&tenantDomain.APIKey{},
		&customerDomain.Customer{},
		&customerDomain.EventLog{},
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
}

func (a *App) setupRouter() {
	a.router = gin.Default()
	a.router.Use(middleware.CORS())
	a.router.Use(middleware.Logger())

	// Health check
	a.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := a.router.Group("/api/v1")
	a.setupModules(v1)
}

func (a *App) setupModules(api *gin.RouterGroup) {
	// Initialize Repositories
	tenantRepository := tenantRepo.NewTenantPostgresRepo(a.db)
	apiKeyRepository := tenantRepo.NewAPIKeyPostgresRepo(a.db)
	customerRepository := customerRepo.NewCustomerPostgresRepo(a.db)
	customerEventRepository := customerRepo.NewCustomerEventPostgresRepo(a.db)
	tripRepository := tripRepo.NewTripPostgresRepo(a.db)
	segmentRepository := segmentationRepo.NewSegmentPostgresRepo(a.db)
	customerSegmentRepository := segmentationRepo.NewCustomerSegmentPostgresRepo(a.db)
	campaignRepository := campaignRepo.NewCampaignPostgresRepo(a.db)
	voucherRepository := campaignRepo.NewVoucherPostgresRepo(a.db)
	voucherUsageRepository := campaignRepo.NewVoucherUsagePostgresRepo(a.db)
	automationRuleRepository := automationRepo.NewAutomationRulePostgresRepo(a.db)
	automationLogRepository := automationRepo.NewAutomationLogPostgresRepo(a.db)
	notificationRepository := notificationRepo.NewNotificationPostgresRepo(a.db)

	// Initialize Services
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

	// Initialize Handlers
	tenantH := tenantHandler.NewTenantHandler(tenantSvc)
	customerH := customerHandler.NewCustomerHandler(customerSvc)
	tripH := tripHandler.NewTripHandler(tripSvc)
	ingestionH := ingestionHandler.NewIngestionHandler(customerSvc, tripSvc)
	segmentationH := segmentationHandler.NewSegmentationHandler(segmentationSvc)
	campaignH := campaignHandler.NewCampaignHandler(campaignSvc)
	automationH := automationHandler.NewAutomationHandler(automationSvc)
	notificationH := notificationHandler.NewNotificationHandler(notificationSvc)
	analyticsH := analyticsHandler.NewAnalyticsHandler(a.db)

	// --- Register Routes ---

	// Admin routes (no API key needed for tenant management)
	tenantH.RegisterRoutes(api)

	// Protected routes (API key required, with cache)
	apiKeyCache := middleware.NewAPIKeyCache()
	protected := api.Group("")
	protected.Use(middleware.APIKeyAuth(a.db, apiKeyCache))
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

	log.Printf("[Server] Modules loaded: tenant, customer, trip, ingestion, segmentation, campaign, automation, notification, analytics")
}
