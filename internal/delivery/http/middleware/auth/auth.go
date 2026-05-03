package auth

import (
	"context"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/errdefs"
	"github.com/labstack/echo/v5"
)

type contextKey string

const userContextKey contextKey = "user"
const userKey = "user"

func New(jwtManager domain.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return errdefs.ErrUnauthorized("missing Authorization header")
			}
			if len(token) < 7 || token[:7] != "Bearer " {
				return errdefs.ErrUnauthorized("Authorization header must start with 'Bearer '")
			}
			token = token[7:]
			if token == "" {
				return errdefs.ErrUnauthorized("missing token in Authorization header")
			}

			claims, err := jwtManager.VerifyAccessToken(token)
			if err != nil {
				return err
			}

			c.Set(userKey, claims)
			req := c.Request()
			ctx := context.WithValue(req.Context(), userContextKey, claims)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func NewWithSession(jwtManager domain.JWTManager, sessionRepo domain.SessionRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := c.Request().Header.Get("Authorization")
			if token == "" {
				return errdefs.ErrUnauthorized("missing Authorization header")
			}
			if len(token) < 7 || token[:7] != "Bearer " {
				return errdefs.ErrUnauthorized("Authorization header must start with 'Bearer '")
			}
			token = token[7:]
			if token == "" {
				return errdefs.ErrUnauthorized("missing token in Authorization header")
			}

			claims, err := jwtManager.VerifyAccessToken(token)
			if err != nil {
				return err
			}

			storedToken, err := sessionRepo.GetAccessToken(c.Request().Context(), claims.ID)
			if err != nil {
				return errdefs.ErrUnauthorized("invalid session")
			}
			if storedToken != token {
				return errdefs.ErrUnauthorized("invalid session")
			}

			c.Set(userKey, claims)
			req := c.Request()
			ctx := context.WithValue(req.Context(), userContextKey, claims)
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}

func GetUser(c *echo.Context) *domain.JWTClaims {
	claims, ok := c.Get(userKey).(*domain.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserFromContext(ctx context.Context) *domain.JWTClaims {
	claims, ok := ctx.Value(userContextKey).(*domain.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
