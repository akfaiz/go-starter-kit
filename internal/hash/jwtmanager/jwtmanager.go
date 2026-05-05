package jwtmanager

import (
	"errors"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	cerrors "github.com/cockroachdb/errors"
	"github.com/golang-jwt/jwt/v5"
)

type jwtManager struct {
	accessSecret   []byte
	refreshSecret  []byte
	accessExpires  time.Duration
	refreshExpires time.Duration
}

func New(cfg config.JWT) domain.JWTManager {
	return &jwtManager{
		accessSecret:   []byte(cfg.AccessSecret),
		refreshSecret:  []byte(cfg.RefreshSecret),
		accessExpires:  cfg.AccessExpires,
		refreshExpires: cfg.RefreshExpires,
	}
}

func (j *jwtManager) GeneratePairToken(claims *domain.JWTClaims) (*domain.PairToken, error) {
	accessToken, err := j.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.GenerateRefreshToken(claims)
	if err != nil {
		return nil, err
	}

	return &domain.PairToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (j *jwtManager) GenerateAccessToken(claims *domain.JWTClaims) (string, error) {
	if claims == nil {
		return "", errors.New("claims cannot be nil")
	}
	return j.generateToken(claims, j.accessSecret, j.accessExpires)
}

func (j *jwtManager) GenerateRefreshToken(claims *domain.JWTClaims) (string, error) {
	if claims == nil {
		return "", errors.New("claims cannot be nil")
	}
	return j.generateToken(claims, j.refreshSecret, j.refreshExpires)
}

func (j *jwtManager) VerifyAccessToken(token string) (*domain.JWTClaims, error) {
	return j.verifyToken(token, j.accessSecret)
}

func (j *jwtManager) VerifyRefreshToken(token string) (*domain.JWTClaims, error) {
	return j.verifyToken(token, j.refreshSecret)
}

func (j *jwtManager) generateToken(claims *domain.JWTClaims, secret []byte, expiresIn time.Duration) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiresIn))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", cerrors.WithStack(err)
	}
	return signedToken, nil
}

func (j *jwtManager) verifyToken(token string, secret []byte) (*domain.JWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &domain.JWTClaims{}, func(_ *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domain.ErrTokenExpired
		}
		return nil, domain.ErrInvalidToken
	}
	if claims, ok := parsedToken.Claims.(*domain.JWTClaims); ok && parsedToken.Valid {
		return claims, nil
	}
	return nil, domain.ErrInvalidToken
}
