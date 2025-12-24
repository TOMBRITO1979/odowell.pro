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

// PoolStats holds connection pool statistics for monitoring
type PoolStats struct {
	MaxOpenConnections int `json:"max_open"`
	OpenConnections    int `json:"open_connections"`
	InUse              int `json:"in_use"`
	Idle               int `json:"idle"`
	WaitCount          int64 `json:"wait_count"`
	WaitDuration       int64 `json:"wait_duration_ms"`
}

// GetPoolStats returns current connection pool statistics
func GetPoolStats() (*PoolStats, error) {
	if healthDB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	stats := healthDB.Stats()
	return &PoolStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration.Milliseconds(),
	}, nil
}

// validatePoolSettings checks if pool configuration is safe
func validatePoolSettings(sqlDB *sql.DB, maxOpen, maxIdle, numReplicas int) error {
	// Query PostgreSQL max_connections
	var maxConnections int
	row := sqlDB.QueryRow("SHOW max_connections")
	if err := row.Scan(&maxConnections); err != nil {
		return fmt.Errorf("could not query max_connections: %v", err)
	}

	// Calculate safe limit (leave 20% for superuser/maintenance connections)
	safeLimit := int(float64(maxConnections) * 0.8)
	totalFromReplicas := maxOpen * numReplicas

	// Validation checks
	if totalFromReplicas > safeLimit {
		return fmt.Errorf("UNSAFE: %d replicas x %d connections = %d exceeds safe limit of %d (80%% of max_connections=%d). Reduce DB_MAX_OPEN_CONNS or DB_NUM_REPLICAS",
			numReplicas, maxOpen, totalFromReplicas, safeLimit, maxConnections)
	}

	if maxIdle > maxOpen {
		return fmt.Errorf("INEFFICIENT: maxIdleConns (%d) > maxOpenConns (%d). Set DB_MAX_IDLE_CONNS <= DB_MAX_OPEN_CONNS", maxIdle, maxOpen)
	}

	if maxOpen > 100 {
		return fmt.Errorf("HIGH: maxOpenConns=%d is very high. Consider reducing for better resource management", maxOpen)
	}

	log.Printf("Pool validation OK: PostgreSQL max_connections=%d, safe_limit=%d, total_from_replicas=%d",
		maxConnections, safeLimit, totalFromReplicas)

	return nil
}

// getEnvInt returns an environment variable as int, or default if not set/invalid
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ValidSSLModes defines which SSL modes are considered secure for production
var ValidSSLModes = map[string]bool{
	"require":     true,
	"verify-ca":   true,
	"verify-full": true,
}

// Connect establishes connection to PostgreSQL database
func Connect() error {
	env := os.Getenv("ENV")
	isProduction := env != "development"

	// Determinar SSL mode baseado no ambiente
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		// Default to require in production, disable in development
		if !isProduction {
			sslMode = "disable"
		} else {
			sslMode = "require" // TLS obrigatorio em producao para proteger dados sensiveis
		}
	}

	// SECURITY: Fail boot if SSL mode is insecure in production
	if isProduction {
		if sslMode == "disable" || sslMode == "prefer" || sslMode == "allow" {
			return fmt.Errorf("SECURITY FATAL: DB_SSL_MODE='%s' is not allowed in production. Use 'require', 'verify-ca', or 'verify-full'", sslMode)
		}
		if !ValidSSLModes[sslMode] {
			return fmt.Errorf("SECURITY FATAL: DB_SSL_MODE='%s' is not a valid secure mode. Use 'require', 'verify-ca', or 'verify-full'", sslMode)
		}
		log.Printf("SSL mode '%s' validated for production", sslMode)
	} else {
		// Development environment - warn but allow
		if sslMode == "disable" {
			log.Println("SECURITY WARNING: Database SSL is disabled - only acceptable in development!")
		}
	}

	// Get timezone from environment variable, default to America/Sao_Paulo
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "America/Sao_Paulo"
	}

	// Build DSN with optional SSL certificate paths for verify-full mode
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

	// Add SSL certificate paths for verify-ca or verify-full modes
	if sslMode == "verify-ca" || sslMode == "verify-full" {
		if sslRootCert := os.Getenv("DB_SSL_ROOT_CERT"); sslRootCert != "" {
			dsn += fmt.Sprintf(" sslrootcert=%s", sslRootCert)
		}
		if sslCert := os.Getenv("DB_SSL_CERT"); sslCert != "" {
			dsn += fmt.Sprintf(" sslcert=%s", sslCert)
		}
		if sslKey := os.Getenv("DB_SSL_KEY"); sslKey != "" {
			dsn += fmt.Sprintf(" sslkey=%s", sslKey)
		}
		log.Printf("SSL certificates configured for mode '%s'", sslMode)
	}

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
	// Recommended: 2 replicas with 100 PG connections = 25 per replica (with buffer for other connections)
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	connMaxLifetimeSecs := getEnvInt("DB_CONN_MAX_LIFETIME", 3600)
	numReplicas := getEnvInt("DB_NUM_REPLICAS", 2)

	// GUARDRAILS: Validate connection pool settings
	if err := validatePoolSettings(sqlDB, maxOpenConns, maxIdleConns, numReplicas); err != nil {
		log.Printf("POOL WARNING: %v", err)
	}

	// Apply pool settings
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetimeSecs) * time.Second)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute) // Close idle connections after 5 minutes

	log.Printf("Database connected successfully - Pool: maxOpen=%d, maxIdle=%d, lifetime=%ds, replicas=%d, timezone=%s",
		maxOpenConns, maxIdleConns, connMaxLifetimeSecs, numReplicas, tz)

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
