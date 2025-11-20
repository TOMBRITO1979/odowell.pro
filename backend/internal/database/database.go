package database

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes connection to PostgreSQL database
func Connect() error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool for better performance
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)                  // Maximum idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Connection lifetime: 1 hour

	log.Println("Database connected successfully with connection pool")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// SetSchema sets the search path for a specific tenant schema
// Validates schema name to prevent SQL injection
func SetSchema(db *gorm.DB, schema string) *gorm.DB {
	// Validate schema name: only alphanumeric and underscore allowed
	validSchema := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validSchema.MatchString(schema) {
		log.Printf("SECURITY WARNING: Invalid schema name attempted: %s", schema)
		// Return db without changing schema - will use current schema
		return db
	}
	return db.Exec(fmt.Sprintf("SET search_path TO %s", schema))
}

// CreateSchema creates a new schema for a tenant
// Validates schema name to prevent SQL injection
func CreateSchema(schemaName string) error {
	// Validate schema name: only alphanumeric and underscore allowed
	validSchema := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validSchema.MatchString(schemaName) {
		return fmt.Errorf("invalid schema name: %s (only alphanumeric and underscore allowed)", schemaName)
	}
	return DB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
}
