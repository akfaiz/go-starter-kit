package domain

import "context"

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
