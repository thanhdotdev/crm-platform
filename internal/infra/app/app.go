package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vothanh/crm-platform/internal/infra/config"
	"github.com/vothanh/crm-platform/internal/infra/database"
	"gitlab.com/bship1/bship-common-go.git/pkg/zlog"
	"gorm.io/gorm"

	// Analytics
	analyticsHandler "github.com/vothanh/crm-platform/internal/modules/analytics/handler"

	// Automation
	automationDomain "github.com/vothanh/crm-platform/internal/modules/automation/domain"
	automationHandler "github.com/vothanh/crm-platform/internal/modules/automation/handler"

	// Campaign
	campaignDomain "github.com/vothanh/crm-platform/internal/modules/campaign/domain"
	campaignHandler "github.com/vothanh/crm-platform/internal/modules/campaign/handler"

	// Customer
	customerDomain "github.com/vothanh/crm-platform/internal/modules/customer/domain"
	customerHandler "github.com/vothanh/crm-platform/internal/modules/customer/handler"

	// Ingestion
	ingestionDomain "github.com/vothanh/crm-platform/internal/modules/ingestion/domain"
	ingestionHandler "github.com/vothanh/crm-platform/internal/modules/ingestion/handler"

	// Notification
	notificationDomain "github.com/vothanh/crm-platform/internal/modules/notification/domain"
	notificationHandler "github.com/vothanh/crm-platform/internal/modules/notification/handler"

	// Segmentation
	segmentationDomain "github.com/vothanh/crm-platform/internal/modules/segmentation/domain"
	segmentationHandler "github.com/vothanh/crm-platform/internal/modules/segmentation/handler"

	// Tenant
	tenantDomain "github.com/vothanh/crm-platform/internal/modules/tenant/domain"
	tenantHandler "github.com/vothanh/crm-platform/internal/modules/tenant/handler"

	// Trip
	tripDomain "github.com/vothanh/crm-platform/internal/modules/trip/domain"
	tripHandler "github.com/vothanh/crm-platform/internal/modules/trip/handler"

	"github.com/vothanh/crm-platform/internal/shared/middleware"
)

// App manages the lifecycle, routing, and dependency injection of the CRM.
type App struct {
	cfg    *config.Config
	db     *gorm.DB
	router *gin.Engine
	server *http.Server
}

// NewAppWithDependencies creates a new App struct with all handlers injected.
// This is called by Google Wire after it has instantiated all dependencies.
func NewAppWithDependencies(
	cfg *config.Config,
	db *gorm.DB,
	tenantH *tenantHandler.TenantHandler,
	customerH *customerHandler.CustomerHandler,
	tripH *tripHandler.TripHandler,
	ingestionH *ingestionHandler.IngestionHandler,
	segmentationH *segmentationHandler.SegmentationHandler,
	campaignH *campaignHandler.CampaignHandler,
	automationH *automationHandler.AutomationHandler,
	notificationH *notificationHandler.NotificationHandler,
	analyticsH *analyticsHandler.AnalyticsHandler,
) (*App, error) {

	app := &App{
		cfg: cfg,
		db:  db,
	}

	// if err := app.migrate(); err != nil {
	// 	return nil, fmt.Errorf("failed to auto-migrate database: %w", err)
	// }

	app.router = gin.Default()
	app.router.Use(middleware.CORS())
	app.router.Use(middleware.Logger())

	// Health check
	app.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := app.router.Group("/api/v1")

	// Admin routes
	tenantH.RegisterRoutes(api)

	// Protected routes (API key required)
	apiKeyCache := middleware.NewAPIKeyCache()
	protected := api.Group("")
	protected.Use(middleware.APIKeyAuth(app.db, apiKeyCache))
	{
		ingestionH.RegisterRoutes(protected)
		customerH.RegisterRoutes(protected)
		tripH.RegisterRoutes(protected)
		segmentationH.RegisterRoutes(protected)
		campaignH.RegisterRoutes(protected)
		automationH.RegisterRoutes(protected)
		notificationH.RegisterRoutes(protected)
		analyticsH.RegisterRoutes(protected)
	}

	zlog.Info("Modules loaded via Wire: tenant, customer, trip, ingestion, segmentation, campaign, automation, notification, analytics")

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
		zlog.Infof("Starting CRM Platform on %s", addr)
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
		zlog.Infof("Shutdown signal received: %v. Initiating graceful shutdown...", sig)

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
		&ingestionDomain.EventLog{},
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
