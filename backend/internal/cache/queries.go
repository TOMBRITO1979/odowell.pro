package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache TTLs for different types of queries
const (
	TTLDentists     = 5 * time.Minute  // Lista de dentistas
	TTLProtocols    = 10 * time.Minute // Procedimentos/Protocolos
	TTLSettings     = 5 * time.Minute  // Configurações do tenant
	TTLCounts       = 1 * time.Minute  // Contagem de pendências
	TTLPatientList  = 2 * time.Minute  // Lista de pacientes
	TTLAppointments = 1 * time.Minute  // Lista de agendamentos
)

// Cache key prefixes
const (
	PrefixDentists       = "dentists"
	PrefixProtocols      = "protocols"
	PrefixSettings       = "settings"
	PrefixPendingCount   = "pending_count"
	PrefixOverdueCount   = "overdue_count"
	PrefixPatients       = "patients"
	PrefixAppointments   = "appointments"
	PrefixDashboard      = "dashboard"
)

// GetOrSet retrieves a value from cache or fetches it using the provided function
// If Redis is not available, it falls back to calling fetchFunc directly
func GetOrSet(key string, ttl time.Duration, fetchFunc func() (interface{}, error)) (interface{}, error) {
	// Check if Redis is available
	if Client == nil {
		return fetchFunc() // Fallback to direct fetch
	}

	// Try to get from cache
	cached, err := Client.Get(ctx, key).Bytes()
	if err == nil {
		var result interface{}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	// Not in cache or error, fetch from source
	if err != nil && err != redis.Nil {
		// Log Redis error but continue
		// helpers.LogWarn("Redis cache read error", map[string]interface{}{"key": key, "error": err.Error()})
	}

	// Fetch from database
	result, err := fetchFunc()
	if err != nil {
		return nil, err
	}

	// Store in cache (async, don't block on cache write errors)
	go func() {
		data, _ := json.Marshal(result)
		Client.Set(ctx, key, data, ttl)
	}()

	return result, nil
}

// GetOrSetTyped is a generic version of GetOrSet for typed results
// Uses json marshal/unmarshal for type conversion
func GetOrSetTyped[T any](key string, ttl time.Duration, fetchFunc func() (T, error)) (T, error) {
	var result T

	// Check if Redis is available
	if Client == nil {
		return fetchFunc() // Fallback to direct fetch
	}

	// Try to get from cache
	cached, err := Client.Get(ctx, key).Bytes()
	if err == nil {
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	// Fetch from database
	result, err = fetchFunc()
	if err != nil {
		return result, err
	}

	// Store in cache (async)
	go func() {
		data, _ := json.Marshal(result)
		Client.Set(ctx, key, data, ttl)
	}()

	return result, nil
}

// InvalidatePrefix removes all cache entries with a given prefix
func InvalidatePrefix(prefix string) error {
	if Client == nil {
		return nil // No Redis, nothing to invalidate
	}

	pattern := fmt.Sprintf("%s:*", prefix)
	return DeletePattern(pattern)
}

// InvalidateKey removes a specific cache entry
func InvalidateKey(key string) error {
	if Client == nil {
		return nil
	}
	return Delete(key)
}

// InvalidateTenantCache invalidates all cache entries for a tenant
func InvalidateTenantCache(tenantID uint, prefixes ...string) error {
	if Client == nil {
		return nil
	}

	for _, prefix := range prefixes {
		pattern := fmt.Sprintf("%s:%d:*", prefix, tenantID)
		if err := DeletePattern(pattern); err != nil {
			return err
		}
	}
	return nil
}

// Cache key builders for common queries

// DentistsKey returns the cache key for tenant's dentist list
func DentistsKey(tenantID uint) string {
	return CacheKey(PrefixDentists, tenantID)
}

// ProtocolsKey returns the cache key for tenant's protocols
func ProtocolsKey(tenantID uint) string {
	return CacheKey(PrefixProtocols, tenantID)
}

// SettingsKey returns the cache key for tenant settings
func SettingsKey(tenantID uint) string {
	return CacheKey(PrefixSettings, tenantID)
}

// PendingCountKey returns the cache key for user's pending task count
func PendingCountKey(userID uint) string {
	return CacheKey(PrefixPendingCount, userID)
}

// OverdueCountKey returns the cache key for tenant's overdue payment count
func OverdueCountKey(tenantID uint) string {
	return CacheKey(PrefixOverdueCount, tenantID)
}

// DashboardKey returns the cache key for tenant's dashboard data
func DashboardKey(tenantID uint, startDate, endDate string) string {
	return CacheKey(PrefixDashboard, tenantID, startDate, endDate)
}

// Helper function to invalidate common cache patterns

// InvalidateOnUserChange invalidates caches affected by user changes
func InvalidateOnUserChange(tenantID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixDentists, tenantID))
	}()
}

// InvalidateOnProtocolChange invalidates protocol cache
func InvalidateOnProtocolChange(tenantID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixProtocols, tenantID))
	}()
}

// InvalidateOnSettingsChange invalidates settings cache
func InvalidateOnSettingsChange(tenantID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidateKey(SettingsKey(tenantID))
	}()
}

// InvalidateOnTaskChange invalidates task count cache
func InvalidateOnTaskChange(userID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidateKey(PendingCountKey(userID))
	}()
}

// InvalidateOnPaymentChange invalidates payment-related caches
func InvalidateOnPaymentChange(tenantID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidateKey(OverdueCountKey(tenantID))
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixDashboard, tenantID))
	}()
}

// InvalidateOnPatientChange invalidates patient-related caches
func InvalidateOnPatientChange(tenantID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixPatients, tenantID))
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixDashboard, tenantID))
	}()
}

// InvalidateOnAppointmentChange invalidates appointment-related caches
func InvalidateOnAppointmentChange(tenantID uint) {
	if Client == nil {
		return
	}
	go func() {
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixAppointments, tenantID))
		InvalidatePrefix(fmt.Sprintf("%s:%d", PrefixDashboard, tenantID))
	}()
}
