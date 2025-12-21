package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var healthDB *sql.DB // sql.DB for health checks

// getEnvInt returns an environment variable as int, or default if not set/invalid
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Connect establishes connection to PostgreSQL database
func Connect() error {
	// Determinar SSL mode baseado no ambiente
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		// Default to require in production, disable in development
		if os.Getenv("ENV") == "development" {
			sslMode = "disable"
		} else {
			sslMode = "prefer"
		}
	}

	// Get timezone from environment variable, default to America/Sao_Paulo
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "America/Sao_Paulo"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		sslMode,
		tz,
	)

	// Configure logger based on environment with slow query logging
	logLevel := logger.Warn
	if os.Getenv("ENV") == "development" {
		logLevel = logger.Info
	}

	// Custom logger configuration with slow query threshold
	// Queries taking longer than 200ms will be logged as warnings
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, // Log queries slower than 200ms
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,  // Don't log ErrRecordNotFound
			Colorful:                  false, // Disable colors in production logs
		},
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 customLogger,
		SkipDefaultTransaction: true, // Better performance for read-heavy workloads
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool for better performance
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// Connection pool settings - configurable via environment variables
	// For horizontal scaling: each replica should use maxOpenConns = totalPgConns / numReplicas
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 50)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetimeSecs := getEnvInt("DB_CONN_MAX_LIFETIME", 3600)

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetimeSecs) * time.Second)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute) // Close idle connections after 5 minutes

	log.Printf("Database connected successfully - Pool: maxOpen=%d, maxIdle=%d, lifetime=%ds, timezone=%s",
		maxOpenConns, maxIdleConns, connMaxLifetimeSecs, tz)

	// Use the same sql.DB from GORM for health checks
	log.Println("Setting up health check connection...")
	healthDB = sqlDB // Use the same sqlDB we already have
	log.Println("Health check connection configured")

	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Health checks database connectivity using dedicated connection
func Health() error {
	if healthDB == nil {
		return fmt.Errorf("health check connection not initialized")
	}
	return healthDB.Ping()
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

// TenantDB wraps a gorm.DB with tenant-specific schema
type TenantDB struct {
	DB *gorm.DB
}

// GetTenantDBByID returns a database connection configured for a specific tenant
func GetTenantDBByID(tenantID uint) *TenantDB {
	schemaName := fmt.Sprintf("tenant_%d", tenantID)
	db := SetSchema(DB.Session(&gorm.Session{}), schemaName)
	return &TenantDB{DB: db}
}
