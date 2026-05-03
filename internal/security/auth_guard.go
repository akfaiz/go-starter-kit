//go:generate mockgen -source=auth_guard.go -destination=../mocks/security_mock.go -package=mocks
package security

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/redis/go-redis/v9"
)

type RateLimitResult struct {
	Limited       bool
	RetryAfter    int
	LockoutActive bool
}

type AuthGuard interface {
	CheckLogin(ctx context.Context, ip, email string) (*RateLimitResult, error)
	OnLoginFailure(ctx context.Context, ip, email string) (*RateLimitResult, error)
	OnLoginSuccess(ctx context.Context, email string) error
	CheckRefresh(ctx context.Context, ip string) (*RateLimitResult, error)
}

type authGuard struct {
	cfg config.RateLimit
	rdc config.Redis
	rdb *redis.Client
}

func NewAuthGuard(cfg config.Config, rdb *redis.Client) AuthGuard {
	return &authGuard{cfg: cfg.RateLimit, rdc: cfg.Redis, rdb: rdb}
}

func (g *authGuard) CheckLogin(ctx context.Context, ip, email string) (*RateLimitResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return &RateLimitResult{}, nil
	}
	lockKey := g.key("auth:login:lock", email)
	ttl, err := g.rdb.TTL(ctx, lockKey).Result()
	if err != nil {
		return nil, err
	}
	if ttl > 0 {
		return &RateLimitResult{Limited: true, RetryAfter: int(ttl.Seconds()), LockoutActive: true}, nil
	}
	return g.checkWindowLimit(ctx, g.key("auth:login:ip", ip), g.cfg.LoginAttempts, g.cfg.LoginWindow)
}

func (g *authGuard) OnLoginFailure(ctx context.Context, ip, email string) (*RateLimitResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return &RateLimitResult{}, nil
	}

	attemptKey := g.key("auth:login:email:attempt", email)
	pipe := g.rdb.TxPipeline()
	attemptsCmd := pipe.Incr(ctx, attemptKey)
	pipe.Expire(ctx, attemptKey, g.cfg.LoginWindow)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	if attemptsCmd.Val() >= g.cfg.LoginLockoutThreshold {
		lockKey := g.key("auth:login:lock", email)
		if err := g.rdb.Set(ctx, lockKey, 1, g.cfg.LoginLockoutDuration).Err(); err != nil {
			return nil, err
		}
		return &RateLimitResult{
			Limited:       true,
			RetryAfter:    int(g.cfg.LoginLockoutDuration.Seconds()),
			LockoutActive: true,
		}, nil
	}
	return g.CheckLogin(ctx, ip, email)
}

func (g *authGuard) OnLoginSuccess(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil
	}
	keys := []string{
		g.key("auth:login:email:attempt", email),
		g.key("auth:login:lock", email),
	}
	return g.rdb.Del(ctx, keys...).Err()
}

func (g *authGuard) CheckRefresh(ctx context.Context, ip string) (*RateLimitResult, error) {
	return g.checkWindowLimit(ctx, g.key("auth:refresh:ip", ip), g.cfg.RefreshAttemptsPerIP, g.cfg.RefreshWindow)
}

func (g *authGuard) checkWindowLimit(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (*RateLimitResult, error) {
	pipe := g.rdb.TxPipeline()
	countCmd := pipe.Incr(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	pipe.ExpireNX(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	count := countCmd.Val()
	ttl := ttlCmd.Val()
	if ttl <= 0 {
		ttl = window
	}
	if count > limit {
		return &RateLimitResult{Limited: true, RetryAfter: int(ttl.Seconds())}, nil
	}
	return &RateLimitResult{}, nil
}

func (g *authGuard) key(parts ...string) string {
	return fmt.Sprintf("%s:%s", g.rdc.Prefix, strings.Join(parts, ":"))
}
