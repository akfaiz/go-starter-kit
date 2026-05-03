package model

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/uptrace/bun"
)

type User struct {
	ID              int64      `bun:"id,pk,autoincrement"`
	Name            string     `bun:"name,notnull"`
	Email           string     `bun:"email,unique,notnull"`
	Password        string     `bun:"password,notnull"`
	EmailVerifiedAt *time.Time `bun:"email_verified_at"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:current_timestamp"`
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

func ApplyUserUpdate(query *bun.UpdateQuery, update *domain.UserUpdate) *bun.UpdateQuery {
	if update.Name.IsValue() {
		query = query.Set("name = ?", update.Name.MustGet())
	}
	if update.Email.IsValue() {
		query = query.Set("email = ?", update.Email.MustGet())
	}
	if update.Password.IsValue() {
		query = query.Set("password = ?", update.Password.MustGet())
	}
	if update.EmailVerifiedAt.IsValue() {
		if update.EmailVerifiedAt.IsNull() {
			query = query.Set("email_verified_at = NULL")
		} else {
			t := update.EmailVerifiedAt.MustGet()
			query = query.Set("email_verified_at = ?", t)
		}
	}
	return query.Set("updated_at = NOW()")
}
