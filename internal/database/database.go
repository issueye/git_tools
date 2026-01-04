package database

import (
	"fmt"
	"os"
	"path/filepath"

	"git-ai-tools/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Init initializes the database connection
func Init() error {
	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	configDir = filepath.Join(configDir, "git-ai-tools")
	os.MkdirAll(configDir, 0755)

	dbPath := filepath.Join(configDir, "data.db")

	// Open database connection
	var dbErr error
	db, dbErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if dbErr != nil {
		return fmt.Errorf("failed to connect to database: %w", dbErr)
	}

	// Run migrations
	if err := migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// migrate runs database migrations
func migrate() error {
	return db.AutoMigrate(
		&models.RepositoryDB{},
		&models.PromptDB{},
		&models.CommandDB{},
		&models.AppConfigDB{},
		&models.RecentRepoDB{},
	)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}

// Close closes the database connection
func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
