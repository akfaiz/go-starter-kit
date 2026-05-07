package dto

import (
	"time"

	"github.com/akfaiz/go-starter-kit/internal/domain"
)

type UserListRequest struct {
	PaginationRequest
}

type UserGetRequest struct {
	ID int64 `param:"id" validate:"required" label:"User ID"`
}

type UserResponse struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func NewUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}
