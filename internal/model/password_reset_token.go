package model

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
)

type PasswordResetToken struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	UserID    int64     `gorm:"uniqueIndex;not null"`
	Token     string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
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
