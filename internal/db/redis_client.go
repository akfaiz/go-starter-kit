package db

import (
	"context"
	"fmt"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("instrument redis tracing: %w", err)
	}
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		return nil, fmt.Errorf("instrument redis metrics: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
