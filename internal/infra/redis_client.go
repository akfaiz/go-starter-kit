package infra

import (
	"github.com/akfaiz/go-starter-kit/internal/config"
	cerrors "github.com/cockroachdb/errors"
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
		return nil, cerrors.Wrap(err, "instrument redis tracing")
	}
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		return nil, cerrors.Wrap(err, "instrument redis metrics")
	}

	return rdb, nil
}
