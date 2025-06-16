package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"aycom/backend/api-gateway/config"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// InitRedis initializes the Redis client
func InitRedis(cfg *config.Config) error {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis client initialized successfully")
	return nil
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return redisClient
}

// SetCache stores data in Redis with TTL
func SetCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = redisClient.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	log.Printf("Cache set for key: %s, TTL: %v", key, ttl)
	return nil
}

// GetCache retrieves data from Redis
func GetCache(ctx context.Context, key string, dest interface{}) error {
	if redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	data, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss for key: %s", key)
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	log.Printf("Cache hit for key: %s", key)
	return nil
}

// DeleteCache removes data from Redis
func DeleteCache(ctx context.Context, key string) error {
	if redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	err := redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	log.Printf("Cache deleted for key: %s", key)
	return nil
}

// DeleteCachePattern removes all keys matching a pattern
func DeleteCachePattern(ctx context.Context, pattern string) error {
	if redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	keys, err := redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys by pattern: %w", err)
	}

	if len(keys) > 0 {
		err = redisClient.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
		log.Printf("Deleted %d cache keys matching pattern: %s", len(keys), pattern)
	}

	return nil
}
