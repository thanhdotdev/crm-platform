package database

import (
	"fmt" // This import is no longer needed if zlog is used for all logging
	"time"

	"gitlab.com/bship1/bship-common-go.git/pkg/zlog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB creates and returns a GORM database connection to PostgreSQL.
func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
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

	zlog.Info("Connected to PostgreSQL successfully (pool: 50 max, 10 idle)")
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
