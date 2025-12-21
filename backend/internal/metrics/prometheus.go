package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request counters
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request duration histogram
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// HTTP requests in flight
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	// HTTP request errors
	HTTPRequestErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_errors_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"method", "path", "error_type"},
	)

	// Database query counters
	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	// Database query duration histogram
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"operation"},
	)

	// Database connection pool stats
	DBConnectionsOpen = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_open",
			Help: "Current number of open database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Current number of idle database connections",
		},
	)

	DBConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Current number of database connections in use",
		},
	)

	// Business metrics
	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Number of active users (logged in within last 24h)",
		},
	)

	TotalTenants = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_tenants",
			Help: "Total number of active tenants",
		},
	)

	// Login attempts
	LoginAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "login_attempts_total",
			Help: "Total number of login attempts",
		},
		[]string{"success"},
	)

	// 2FA usage
	TwoFactorVerifications = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "two_factor_verifications_total",
			Help: "Total number of 2FA verifications",
		},
		[]string{"success"},
	)
)

// PrometheusMiddleware collects HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint to avoid recursion
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Track requests in flight
		HTTPRequestsInFlight.Inc()
		defer HTTPRequestsInFlight.Dec()

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Normalize path for cardinality control
		path := normalizePath(c.FullPath())
		if path == "" {
			path = "unknown"
		}

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		HTTPRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)

		// Track errors
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			HTTPRequestErrors.WithLabelValues(c.Request.Method, path, errorType).Inc()
		}
	}
}

// normalizePath reduces cardinality by replacing IDs with placeholders
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	// The FullPath from gin already has placeholders like :id
	return path
}

// MetricsHandler returns the Prometheus metrics handler
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordDBQuery records a database query metric
func RecordDBQuery(operation, table string, duration time.Duration) {
	DBQueriesTotal.WithLabelValues(operation, table).Inc()
	DBQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// UpdateDBPoolStats updates the database connection pool metrics
func UpdateDBPoolStats(open, idle, inUse int) {
	DBConnectionsOpen.Set(float64(open))
	DBConnectionsIdle.Set(float64(idle))
	DBConnectionsInUse.Set(float64(inUse))
}

// RecordLogin records a login attempt
func RecordLogin(success bool) {
	LoginAttempts.WithLabelValues(strconv.FormatBool(success)).Inc()
}

// RecordTwoFactorVerification records a 2FA verification attempt
func RecordTwoFactorVerification(success bool) {
	TwoFactorVerifications.WithLabelValues(strconv.FormatBool(success)).Inc()
}
