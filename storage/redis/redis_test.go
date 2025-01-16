package redis_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	rds "github.com/deasdania/dating-app/storage/redis"
	"github.com/redis/go-redis/v9"
)

var testRedisCl *rds.RedisConnection

func TestMain(m *testing.M) {
	const connStr = "REDIS_CONNECTION"
	redisConnStr := os.Getenv(connStr)
	if redisConnStr == "" {
		log.Printf("%s is not set, skipping", connStr)
		os.Exit(1)
	}

	redisConnStr = normalizeRedisURL(redisConnStr, 15)
	var cleanUpFn func()
	var err error
	testRedisCl, cleanUpFn, err = buildConnection(redisConnStr)
	if err != nil {
		log.Printf("failed to build DB connection: %v", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	if cleanUpFn != nil {
		cleanUpFn()
	}

	os.Exit(exitCode)
}

func normalizeRedisURL(url string, db int) string {
	re := regexp.MustCompile(`/([0-9]|1[0-5])$`)
	url = re.ReplaceAllString(url, "")
	return fmt.Sprintf("%s/%d", url, db)
}

func buildConnection(connStr string) (*rds.RedisConnection, func(), error) {
	options, err := redis.ParseURL(connStr)
	if err != nil {
		log.Fatalf("Could not parse Redis URL: %v", err)
	}

	ctx := context.Background()
	rdb := redis.NewClient(options)

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	cleanUpFn := func() {
		err := rdb.FlushDB(ctx).Err()
		if err != nil {
			log.Printf("Could not flush Redis database: %v", err)
		}

		err = rdb.Close()
		if err != nil {
			log.Printf("Could not close Redis connection: %v", err)
		}
	}

	return &rds.RedisConnection{Cl: rdb}, cleanUpFn, nil
}

func TestRedisConnection(t *testing.T) {
	// Use the testRedisCl client in your tests
	_, err := testRedisCl.Cl.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("failed to ping Redis: %v", err)
	}

	// Your test logic here
	t.Log("Successfully connected to Redis")
}
