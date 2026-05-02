//go:generate mockgen -source=user.go -destination=../mocks/user_mock.go -package=mocks
package domain

import (
	"context"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int64) (*User, error)
	Update(ctx context.Context, id int64, user *UserUpdate) error
	Delete(ctx context.Context, id int64) error
}

type UserService interface {
	FindByID(ctx context.Context, id int64) (*User, error)
	UpdateProfile(ctx context.Context, id int64, user *User) error
	ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error
	Delete(ctx context.Context, id int64, password string) error
}

type User struct {
	ID              int64
	Name            string
	Email           string
	Password        string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UserUpdate struct {
	Name            omit.Val[string]
	Email           omit.Val[string]
	Password        omit.Val[string]
	EmailVerifiedAt omitnull.Val[time.Time]
}

func (uu *UserUpdate) IsEmpty() bool {
	return uu.Name.IsUnset() && uu.Email.IsUnset() && uu.Password.IsUnset() && uu.EmailVerifiedAt.IsUnset()
}

func (u *User) IsVerified() bool {
	return u.EmailVerifiedAt != nil
}
