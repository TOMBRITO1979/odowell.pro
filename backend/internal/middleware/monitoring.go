package middleware

import (
	"drcrwell/backend/internal/cache"
	"drcrwell/backend/internal/database"
	"log"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// Metrics holds application metrics for monitoring
type Metrics struct {
	RequestCount      int64
	ErrorCount        int64   // 5xx errors
	WarningCount      int64   // 4xx errors
	TotalLatencyMs    int64
	MaxLatencyMs      int64
	SlowRequestCount  int64   // Requests > 1s
	LastErrorTime     int64   // Unix timestamp
	LastAlertTime     int64   // Unix timestamp
}

var (
	appMetrics  Metrics
	alertConfig = AlertConfig{
		ErrorRateThreshold:      0.05,  // 5% error rate
		SlowRequestThreshold:    1000,  // 1 second in ms
		PoolUsageThreshold:      0.80,  // 80% pool usage
		LatencyP99ThresholdMs:   2000,  // 2 seconds
		AlertCooldownSeconds:    300,   // 5 minutes between alerts
	}
)

// AlertConfig holds thresholds for alerting
type AlertConfig struct {
	ErrorRateThreshold      float64
	SlowRequestThreshold    int64 // milliseconds
	PoolUsageThreshold      float64
	LatencyP99ThresholdMs   int64
	AlertCooldownSeconds    int64
}

// MonitoringMiddleware tracks request metrics and triggers alerts
func MonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latencyMs := time.Since(start).Milliseconds()

		// Update metrics atomically
		atomic.AddInt64(&appMetrics.RequestCount, 1)
		atomic.AddInt64(&appMetrics.TotalLatencyMs, latencyMs)

		// Track max latency
		for {
			current := atomic.LoadInt64(&appMetrics.MaxLatencyMs)
			if latencyMs <= current {
				break
			}
			if atomic.CompareAndSwapInt64(&appMetrics.MaxLatencyMs, current, latencyMs) {
				break
			}
		}

		// Track slow requests
		if latencyMs > alertConfig.SlowRequestThreshold {
			atomic.AddInt64(&appMetrics.SlowRequestCount, 1)
		}

		// Track errors
		status := c.Writer.Status()
		if status >= 500 {
			atomic.AddInt64(&appMetrics.ErrorCount, 1)
			atomic.StoreInt64(&appMetrics.LastErrorTime, time.Now().Unix())

			// Log error for alerting
			log.Printf("ALERT: 5xx error on %s %s - Status: %d - Latency: %dms",
				c.Request.Method, c.Request.URL.Path, status, latencyMs)
		} else if status >= 400 {
			atomic.AddInt64(&appMetrics.WarningCount, 1)
		}

		// Check for alert conditions (async)
		go checkAlerts()
	}
}

// checkAlerts evaluates metrics and triggers alerts if thresholds are exceeded
func checkAlerts() {
	now := time.Now().Unix()
	lastAlert := atomic.LoadInt64(&appMetrics.LastAlertTime)

	// Cooldown check
	if now - lastAlert < alertConfig.AlertCooldownSeconds {
		return
	}

	requestCount := atomic.LoadInt64(&appMetrics.RequestCount)
	if requestCount < 100 {
		return // Not enough data
	}

	errorCount := atomic.LoadInt64(&appMetrics.ErrorCount)
	errorRate := float64(errorCount) / float64(requestCount)

	// Check error rate
	if errorRate > alertConfig.ErrorRateThreshold {
		if atomic.CompareAndSwapInt64(&appMetrics.LastAlertTime, lastAlert, now) {
			log.Printf("CRITICAL ALERT: Error rate %.2f%% exceeds threshold %.2f%%",
				errorRate*100, alertConfig.ErrorRateThreshold*100)
		}
		return
	}

	// Check connection pool
	if stats, err := database.GetPoolStats(); err == nil && stats != nil {
		if stats.MaxOpenConnections > 0 {
			poolUsage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
			if poolUsage > alertConfig.PoolUsageThreshold {
				if atomic.CompareAndSwapInt64(&appMetrics.LastAlertTime, lastAlert, now) {
					log.Printf("CRITICAL ALERT: DB pool usage %.2f%% exceeds threshold %.2f%% (InUse: %d, Max: %d)",
						poolUsage*100, alertConfig.PoolUsageThreshold*100, stats.InUse, stats.MaxOpenConnections)
				}
				return
			}
		}
	}
}

// GetMetrics returns current application metrics
func GetMetrics() map[string]interface{} {
	requestCount := atomic.LoadInt64(&appMetrics.RequestCount)
	errorCount := atomic.LoadInt64(&appMetrics.ErrorCount)
	warningCount := atomic.LoadInt64(&appMetrics.WarningCount)
	totalLatency := atomic.LoadInt64(&appMetrics.TotalLatencyMs)
	slowRequests := atomic.LoadInt64(&appMetrics.SlowRequestCount)
	maxLatency := atomic.LoadInt64(&appMetrics.MaxLatencyMs)

	var avgLatency float64
	var errorRate float64
	if requestCount > 0 {
		avgLatency = float64(totalLatency) / float64(requestCount)
		errorRate = float64(errorCount) / float64(requestCount) * 100
	}

	// Get pool stats
	var poolStats interface{}
	if stats, err := database.GetPoolStats(); err == nil {
		poolStats = stats
	}

	// Get Redis stats
	var redisConnected bool
	if cache.GetClient() != nil {
		redisConnected = cache.Health() == nil
	}

	return map[string]interface{}{
		"requests": map[string]interface{}{
			"total":         requestCount,
			"errors_5xx":    errorCount,
			"warnings_4xx":  warningCount,
			"slow_requests": slowRequests,
			"error_rate":    errorRate,
		},
		"latency": map[string]interface{}{
			"avg_ms": avgLatency,
			"max_ms": maxLatency,
		},
		"db_pool":         poolStats,
		"redis_connected": redisConnected,
		"alerts": map[string]interface{}{
			"last_error_time": atomic.LoadInt64(&appMetrics.LastErrorTime),
			"last_alert_time": atomic.LoadInt64(&appMetrics.LastAlertTime),
		},
	}
}

// ResetMetrics resets all metrics (useful for testing or periodic reset)
func ResetMetrics() {
	atomic.StoreInt64(&appMetrics.RequestCount, 0)
	atomic.StoreInt64(&appMetrics.ErrorCount, 0)
	atomic.StoreInt64(&appMetrics.WarningCount, 0)
	atomic.StoreInt64(&appMetrics.TotalLatencyMs, 0)
	atomic.StoreInt64(&appMetrics.MaxLatencyMs, 0)
	atomic.StoreInt64(&appMetrics.SlowRequestCount, 0)
}
