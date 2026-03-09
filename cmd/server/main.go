package main

import (
	"github.com/vothanh/crm-platform/internal/infra/app"
	"github.com/vothanh/crm-platform/internal/infra/config"
	"gitlab.com/bship1/bship-common-go.git/pkg/zlog"
	"gitlab.com/bship1/bship-common-go.git/pkg/zserver"
)

func main() {
	zserver.Init()
	defer zserver.Shutdown()

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		zlog.Fatal("Failed to load config", zlog.Field("error", err))
	}

	// 2. Initialize application (DI, Router, DB)
	application, err := app.InitializeApp(cfg)
	if err != nil {
		zlog.Fatal("Failed to initialize app", zlog.Field("error", err))
	}

	// 3. Run application (Block and wait for shutdown)
	if err := application.Run(); err != nil {
		zlog.Fatal("Application stopped with error", zlog.Field("error", err))
	}

	zlog.Info("Graceful shutdown completed cleanly.")
}
