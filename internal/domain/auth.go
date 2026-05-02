//go:generate mockgen -source=auth.go -destination=../mocks/auth_mock.go -package=mocks
package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Register(ctx context.Context, user *User) (*PairToken, error)
	Login(ctx context.Context, email, password string) (*PairToken, error)
	RefreshToken(ctx context.Context, refreshToken string) (*PairToken, error)
	SendForgotPasswordOTP(ctx context.Context, email string) error
	VerifyForgotPasswordOTP(ctx context.Context, email, otp string) error
	ResetPasswordWithOTP(ctx context.Context, email, otp, newPassword string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hashed string) (bool, error)
}

type JWTManager interface {
	GeneratePairToken(claims *JWTClaims) (*PairToken, error)
	GenerateAccessToken(claims *JWTClaims) (string, error)
	GenerateRefreshToken(claims *JWTClaims) (string, error)
	VerifyAccessToken(token string) (*JWTClaims, error)
	VerifyRefreshToken(token string) (*JWTClaims, error)
}

type JWTClaims struct {
	jwt.RegisteredClaims
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PairToken struct {
	AccessToken  string
	RefreshToken string
}
