//go:generate mockgen -source=session.go -destination=../mocks/session_mock.go -package=mocks
package domain

import (
	"context"
	"time"
)

type SessionRepository interface {
	SavePairToken(
		ctx context.Context,
		userID int64,
		accessToken, refreshToken string,
		accessTTL, refreshTTL time.Duration,
	) error
	GetAccessToken(ctx context.Context, userID int64) (string, error)
	GetRefreshToken(ctx context.Context, userID int64) (string, error)
	DeleteSession(ctx context.Context, userID int64) error
}
