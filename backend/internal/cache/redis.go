package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

// Connect establishes connection to Redis
func Connect() error {
	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	Client = redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password:        os.Getenv("REDIS_PASSWORD"),
		DB:              redisDB,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 500 * time.Millisecond,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        100,
		MinIdleConns:    10,
		PoolTimeout:     4 * time.Second,
	})

	// Test connection
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Printf("Redis connected successfully at %s:%s", redisHost, redisPort)
	return nil
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return Client
}

// Set stores a value in Redis with expiration
func Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}
	return Client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from Redis
func Get(key string, dest interface{}) error {
	data, err := Client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a key from Redis
func Delete(key string) error {
	return Client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func DeletePattern(pattern string) error {
	iter := Client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := Client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Exists checks if a key exists
func Exists(key string) (bool, error) {
	count, err := Client.Exists(ctx, key).Result()
	return count > 0, err
}

// SetNX sets a value only if the key does not exist (for locks)
func SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %v", err)
	}
	return Client.SetNX(ctx, key, data, expiration).Result()
}

// Incr increments a counter
func Incr(key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

// Expire sets expiration on a key
func Expire(key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// Close closes the Redis connection
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Health checks if Redis is healthy
func Health() error {
	_, err := Client.Ping(ctx).Result()
	return err
}

// CacheKey generates a consistent cache key with prefix
func CacheKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key = fmt.Sprintf("%s:%v", key, part)
	}
	return key
}

// Cache durations
const (
	CacheShort  = 5 * time.Minute
	CacheMedium = 30 * time.Minute
	CacheLong   = 24 * time.Hour
)

// Refresh token durations
const (
	AccessTokenExpiry  = 15 * time.Minute // Short-lived access token
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)

// RefreshTokenData stores refresh token metadata
type RefreshTokenData struct {
	UserID   uint   `json:"user_id"`
	TenantID uint   `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsSuperAdmin bool `json:"is_super_admin"`
}

// StoreRefreshToken stores a refresh token in Redis
func StoreRefreshToken(token string, data RefreshTokenData) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	key := CacheKey("refresh_token", token)
	return Set(key, data, RefreshTokenExpiry)
}

// GetRefreshToken retrieves refresh token data from Redis
func GetRefreshToken(token string) (*RefreshTokenData, error) {
	if Client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}
	key := CacheKey("refresh_token", token)
	var data RefreshTokenData
	err := Get(key, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// DeleteRefreshToken removes a refresh token from Redis
func DeleteRefreshToken(token string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	key := CacheKey("refresh_token", token)
	return Delete(key)
}

// DeleteAllUserRefreshTokens removes all refresh tokens for a user (for password change, etc.)
func DeleteAllUserRefreshTokens(userID uint) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	// Note: This is a simple implementation. For production with many tokens,
	// consider using a secondary index or user-specific token sets
	return DeletePattern(fmt.Sprintf("refresh_token:*"))
}
