package model

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
)

type UserToken struct {
	ID        int64     `bun:"id,pk,autoincrement"`
	UserID    int64     `bun:"user_id,notnull"`
	Token     string    `bun:"token,notnull"`
	TokenType string    `bun:"token_type,notnull"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}

func (ut *UserToken) ToDomain() *domain.UserToken {
	return &domain.UserToken{
		ID:        ut.ID,
		UserID:    ut.UserID,
		Token:     ut.Token,
		TokenType: domain.TokenType(ut.TokenType),
		ExpiresAt: ut.ExpiresAt,
		CreatedAt: ut.CreatedAt,
	}
}
