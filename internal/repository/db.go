package repository

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"insider-challenge/pkg/config"
	"insider-challenge/pkg/domain"
)

const (
	maxIdleConns    = 10
	maxOpenConns    = 100
	connMaxLifetime = 1 * time.Hour
)

// InitDB initialize the db connection
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	// Configure logger
	gormLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	// Auto migrate database schema
	if err := db.AutoMigrate(&domain.Message{}); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	// Initialize sample data if database is empty
	// @todo: remove this if working production
	if err := InitSampleData(db); err != nil {
		return nil, fmt.Errorf("initialize sample data: %w", err)
	}

	return db, nil
}
