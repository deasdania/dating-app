package redis

import (
	"context"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Function to set a cache object with expiration
func (rc *RedisConnection) SetCacheObject(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	// Set the value in Redis
	err := rc.Cl.Set(ctx, key, value, duration).Err()
	if err != nil {
		return fmt.Errorf("could not set cache object: %v", err)
	}

	return nil
}

// Function to get a cache object from Redis
func (rc *RedisConnection) GetCacheObject(ctx context.Context, key string) (string, error) {
	// Get the value from Redis
	val, err := rc.Cl.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key does not exist")
	} else if err != nil {
		return "", fmt.Errorf("could not get cache object: %v", err)
	}

	return val, nil
}

// ReplaceCacheObject replaces (or sets) a cache object with expiration
func (rc *RedisConnection) ReplaceCacheObject(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	// Replacing is essentially the same as setting in Redis, so we use Set.
	err := rc.Cl.Set(ctx, key, value, duration).Err()
	if err != nil {
		return fmt.Errorf("could not replace cache object: %v", err)
	}
	return nil
}

// DeleteCacheObject deletes a cache object by key
func (rc *RedisConnection) DeleteCacheObject(ctx context.Context, key string) error {
	err := rc.Cl.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("could not delete cache object: %v", err)
	}
	return nil
}

// RenewCacheObjectTimeout renews or extends the timeout (expiration) of an existing cache object
func (rc *RedisConnection) RenewCacheObjectTimeout(ctx context.Context, key string, duration time.Duration) error {
	// Use EXPIRE to renew the timeout
	err := rc.Cl.Expire(ctx, key, duration).Err()
	if err != nil {
		return fmt.Errorf("could not renew cache object timeout: %v", err)
	}
	return nil
}
