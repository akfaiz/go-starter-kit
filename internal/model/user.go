package model

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
)

type User struct {
	ID              int64      `gorm:"primaryKey;autoIncrement"`
	Name            string     `gorm:"not null"`
	Email           string     `gorm:"uniqueIndex;not null"`
	Password        string     `gorm:"not null"`
	EmailVerifiedAt *time.Time `gorm:"index"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt       time.Time  `gorm:"not null"`
}

func NewUserFromDomain(u *domain.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Password:        u.Password,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

func (u *User) ToDomain() *domain.User {
	return &domain.User{
		ID:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Password:        u.Password,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}
