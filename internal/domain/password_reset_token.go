//go:generate mockgen -source=password_reset_token.go -destination=../../test/mocks/password_reset_token_mock.go -package=mocks
package domain

import (
	"context"
	"time"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *PasswordResetToken) error
	FindOne(ctx context.Context, userID int64) (*PasswordResetToken, error)
	Delete(ctx context.Context, userID int64) error
}

type PasswordResetToken struct {
	ID        int64
	UserID    int64
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
