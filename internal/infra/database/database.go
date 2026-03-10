package database

import (
	"fmt" // This import is no longer needed if zlog is used for all logging
	"time"

	"github.com/vothanh/crm-platform/internal/infra/config"
	"gitlab.com/bship1/bship-common-go.git/pkg/zlog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates and returns a GORM database connection based on config.
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	var gormDialector gorm.Dialector

	if cfg.Database.Driver == "mysql" {
		gormDialector = mysql.Open(cfg.Database.MySQLDSN())
	} else {
		// default to postgres
		gormDialector = postgres.Open(cfg.Database.PostgresDSN())
	}

	db, err := gorm.Open(gormDialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool config for high throughput
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	zlog.Infof("Connected to %s successfully (pool: 50 max, 10 idle)", cfg.Database.Driver)
	return db, nil
}

// AutoMigrate runs GORM auto-migration for all provided models.
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}
	zlog.Info("Auto-migration completed")
	return nil
}
