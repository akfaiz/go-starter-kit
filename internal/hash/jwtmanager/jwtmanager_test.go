package jwtmanager_test

import (
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/internal/hash/jwtmanager"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GeneratePairToken(t *testing.T) {
	cfg := config.JWT{
		AccessSecret:   "access-secret",
		RefreshSecret:  "refresh-secret",
		AccessExpires:  time.Hour,
		RefreshExpires: time.Hour * 24,
	}
	jwtManager := jwtmanager.New(cfg)

	t.Run("should generate pair token successfully", func(t *testing.T) {
		claims := &domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user123",
			},
		}

		pairToken, err := jwtManager.GeneratePairToken(claims)

		require.NoError(t, err)
		assert.NotNil(t, pairToken)
		assert.NotEmpty(t, pairToken.AccessToken)
		assert.NotEmpty(t, pairToken.RefreshToken)
		assert.NotEqual(t, pairToken.AccessToken, pairToken.RefreshToken)
	})

	t.Run("should return error when claims is nil", func(t *testing.T) {
		pairToken, err := jwtManager.GeneratePairToken(nil)

		assert.Error(t, err)
		assert.Nil(t, pairToken)
	})

	t.Run("generated tokens should be verifiable", func(t *testing.T) {
		claims := &domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user456",
			},
		}

		pairToken, err := jwtManager.GeneratePairToken(claims)
		require.NoError(t, err)

		// Verify access token
		accessClaims, err := jwtManager.VerifyAccessToken(pairToken.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, claims.Subject, accessClaims.Subject)

		// Verify refresh token
		refreshClaims, err := jwtManager.VerifyRefreshToken(pairToken.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, claims.Subject, refreshClaims.Subject)
	})
}

func TestJWTManager_VerifyAccessToken(t *testing.T) {
	cfg := config.JWT{
		AccessSecret:   "access-secret",
		RefreshSecret:  "refresh-secret",
		AccessExpires:  time.Hour,
		RefreshExpires: time.Hour * 24,
	}
	jwtManager := jwtmanager.New(cfg)

	t.Run("should verify valid access token successfully", func(t *testing.T) {
		claims := &domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user123",
			},
		}

		token, err := jwtManager.GenerateAccessToken(claims)
		require.NoError(t, err)

		verifiedClaims, err := jwtManager.VerifyAccessToken(token)

		require.NoError(t, err)
		assert.NotNil(t, verifiedClaims)
		assert.Equal(t, claims.Subject, verifiedClaims.Subject)
	})

	t.Run("should return error for malformed token", func(t *testing.T) {
		invalidToken := "invalid.token.here"

		claims, err := jwtManager.VerifyAccessToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should return error for empty token", func(t *testing.T) {
		claims, err := jwtManager.VerifyAccessToken("")

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should return error for token with wrong secret", func(t *testing.T) {
		// Generate token with different manager
		wrongCfg := config.JWT{
			AccessSecret:   "wrong-secret",
			RefreshSecret:  "refresh-secret",
			AccessExpires:  time.Hour,
			RefreshExpires: time.Hour * 24,
		}
		wrongManager := jwtmanager.New(wrongCfg)

		claims := &domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user123",
			},
		}

		token, err := wrongManager.GenerateAccessToken(claims)
		require.NoError(t, err)

		verifiedClaims, err := jwtManager.VerifyAccessToken(token)

		assert.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})

	t.Run("should return error for expired token", func(t *testing.T) {
		expiredCfg := config.JWT{
			AccessSecret:   "access-secret",
			RefreshSecret:  "refresh-secret",
			AccessExpires:  -time.Hour, // Expired token
			RefreshExpires: time.Hour * 24,
		}
		expiredManager := jwtmanager.New(expiredCfg)

		claims := &domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user123",
			},
		}

		token, err := expiredManager.GenerateAccessToken(claims)
		require.NoError(t, err)

		verifiedClaims, err := jwtManager.VerifyAccessToken(token)

		assert.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})

	t.Run("should not verify refresh token as access token", func(t *testing.T) {
		claims := &domain.JWTClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: "user123",
			},
		}

		refreshToken, err := jwtManager.GenerateRefreshToken(claims)
		require.NoError(t, err)

		verifiedClaims, err := jwtManager.VerifyAccessToken(refreshToken)

		assert.Error(t, err)
		assert.Nil(t, verifiedClaims)
	})
}
