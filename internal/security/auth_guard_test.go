package security

import (
	"context"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestGuard(rdb *redis.Client) *authGuard {
	cfg := config.Config{
		Redis: config.Redis{Prefix: "test"},
		RateLimit: config.RateLimit{
			LoginAttempts:         5,
			LoginWindow:           time.Minute,
			LoginLockoutThreshold: 3,
			LoginLockoutDuration:  2 * time.Minute,
			RefreshAttemptsPerIP:  10,
			RefreshWindow:         time.Minute,
		},
	}
	return NewAuthGuard(cfg, rdb).(*authGuard)
}

func TestKey(t *testing.T) {
	g := newTestGuard(nil)
	assert.Equal(t, "test:auth:login:ip:127.0.0.1", g.key("auth", "login", "ip", "127.0.0.1"))
}

func TestLoginPathsIgnoreEmptyEmail(t *testing.T) {
	g := newTestGuard(nil)
	ctx := context.Background()

	res, err := g.CheckLogin(ctx, "127.0.0.1", "   ")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Limited)
	assert.Zero(t, res.RetryAfter)
	assert.False(t, res.LockoutActive)

	res, err = g.OnLoginFailure(ctx, "127.0.0.1", "")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.Limited)
	assert.Zero(t, res.RetryAfter)
	assert.False(t, res.LockoutActive)

	assert.NoError(t, g.OnLoginSuccess(ctx, "   "))
}

func TestCheckRefreshReturnsRedisError(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:0",
		DialTimeout:  20 * time.Millisecond,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
	})
	defer rdb.Close()

	g := newTestGuard(rdb)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	res, err := g.CheckRefresh(ctx, "127.0.0.1")
	assert.Nil(t, res)
	assert.Error(t, err)
}
