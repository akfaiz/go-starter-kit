package session

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/redis/go-redis/v9"
)

type repository struct {
	rdb    *redis.Client
	prefix string
}

func NewRepository(cfg config.Config, rdb *redis.Client) domain.SessionRepository {
	return &repository{
		rdb:    rdb,
		prefix: cfg.Redis.Prefix,
	}
}

func (r *repository) SavePairToken(
	ctx context.Context,
	userID int64,
	accessToken, refreshToken string,
	accessTTL, refreshTTL time.Duration,
) error {
	pipe := r.rdb.TxPipeline()
	pipe.Set(ctx, r.accessKey(userID), accessToken, accessTTL)
	pipe.Set(ctx, r.refreshKey(userID), refreshToken, refreshTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *repository) GetRefreshToken(ctx context.Context, userID int64) (string, error) {
	token, err := r.rdb.Get(ctx, r.refreshKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrResourceNotFound
		}
		return "", err
	}
	return token, nil
}

func (r *repository) GetAccessToken(ctx context.Context, userID int64) (string, error) {
	token, err := r.rdb.Get(ctx, r.accessKey(userID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", domain.ErrResourceNotFound
		}
		return "", err
	}
	return token, nil
}

func (r *repository) DeleteSession(ctx context.Context, userID int64) error {
	return r.rdb.Del(ctx, r.accessKey(userID), r.refreshKey(userID)).Err()
}

func (r *repository) accessKey(userID int64) string {
	return r.key("auth:session", strconv.FormatInt(userID, 10), "access")
}

func (r *repository) refreshKey(userID int64) string {
	return r.key("auth:session", strconv.FormatInt(userID, 10), "refresh")
}

func (r *repository) key(parts ...string) string {
	return fmt.Sprintf("%s:%s", r.prefix, strings.Join(parts, ":"))
}
