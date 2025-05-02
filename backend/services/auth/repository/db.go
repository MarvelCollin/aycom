\
package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormConnection establishes a new database connection using GORM.
func NewGormConnection(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.Username, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	// Configure GORM logger
	newLogger := logger.New(
		log.New(log.Writer(), "\\r\\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: true,        // Don't log ErrRecordNotFound
			Colorful:                  false,       // Disable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Optional: Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("GORM database connection established")
	return db, nil
}

// CreateTables is kept for potential direct use if needed, but AutoMigrate is preferred.
// This function is no longer used by the service layer directly.
func CreateTables(db *gorm.DB) error {
	err := db.AutoMigrate(&User{}, &Token{}, &OAuthConnection{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %w", err)
	}
	log.Println("GORM Auto-migration completed (or tables already exist).")
	return nil
}
