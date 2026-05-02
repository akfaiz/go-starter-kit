package config

import (
	"time"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type RateLimit struct {
	LoginAttempts         int64
	LoginWindow           time.Duration
	LoginLockoutThreshold int64
	LoginLockoutDuration  time.Duration
	RefreshAttemptsPerIP  int64
	RefreshWindow         time.Duration
}

func loadRateLimitConfig() RateLimit {
	return RateLimit{
		LoginAttempts:         int64(env.GetInt("AUTH_LOGIN_LIMIT_ATTEMPTS", 5)),
		LoginWindow:           env.GetDuration("AUTH_LOGIN_LIMIT_WINDOW", 10*time.Minute),
		LoginLockoutThreshold: int64(env.GetInt("AUTH_LOGIN_LOCKOUT_THRESHOLD", 5)),
		LoginLockoutDuration:  env.GetDuration("AUTH_LOGIN_LOCKOUT_DURATION", 15*time.Minute),
		RefreshAttemptsPerIP:  int64(env.GetInt("AUTH_REFRESH_LIMIT_ATTEMPTS", 20)),
		RefreshWindow:         env.GetDuration("AUTH_REFRESH_LIMIT_WINDOW", 10*time.Minute),
	}
}
