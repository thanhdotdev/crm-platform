package main

import (
	"log"

	"github.com/vothanh/crm-platform/internal/app"
	"github.com/vothanh/crm-platform/internal/config"
	"gitlab.com/bship1/bship-common-go.git/pkg/zserver"
)

func main() {
	zserver.Init()
	defer zserver.Shutdown()

	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Main] Failed to load config: %v", err)
	}

	// 2. Initialize application (DI, Router, DB)
	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize app: %v", err)
	}

	// 3. Run application (Block and wait for shutdown)
	if err := application.Run(); err != nil {
		log.Fatalf("[Main] Application stopped with error: %v", err)
	}

	log.Println("[Main] Graceful shutdown completed cleanly.")
}
