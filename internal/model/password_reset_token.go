package model

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
)

type PasswordResetToken struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	UserID    int64     `bun:"user_id,notnull"`
	Token     string    `bun:"token,notnull"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

func NewPasswordResetTokenFromDomain(t *domain.PasswordResetToken) *PasswordResetToken {
	if t == nil {
		return nil
	}
	return &PasswordResetToken{
		ID:        t.ID,
		UserID:    t.UserID,
		Token:     t.Token,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
}

func (ut *PasswordResetToken) ToDomain() *domain.PasswordResetToken {
	return &domain.PasswordResetToken{
		ID:        ut.ID,
		UserID:    ut.UserID,
		Token:     ut.Token,
		ExpiresAt: ut.ExpiresAt,
		CreatedAt: ut.CreatedAt,
	}
}
