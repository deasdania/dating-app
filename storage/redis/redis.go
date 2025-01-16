package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	Password string
	Database int
	Timeout  time.Duration
	SSL      bool
}

type RedisConnection struct {
	Cl *redis.Client
}

func NewRedisConnection(cfg Config) (*RedisConnection, error) {
	redisOptions := &redis.Options{
		Addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:    cfg.Password,
		DB:          cfg.Database,
		DialTimeout: cfg.Timeout,
		TLSConfig:   getTLSConfig(cfg.SSL),
	}

	cl := redis.NewClient(redisOptions)
	if err := cl.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisConnection{
		Cl: cl,
	}, nil
}

// NOTE: we need to configure proper TLS later
func getTLSConfig(sslEnabled bool) *tls.Config {
	if sslEnabled {
		return &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	return nil
}
