package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
)

type contextKey string

const userContextKey contextKey = "user"
const userKey = "user"

// New is a middleware that verifies the JWT token and sets the user information in the context.
func New(jwtManager domain.JWTManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token, err := extractToken(c)
			if err != nil {
				return err
			}

			claims, err := jwtManager.VerifyAccessToken(token)
			if err != nil {
				return handleTokenError(err)
			}

			setUser(c, claims)
			return next(c)
		}
	}
}

// NewWithSession is a middleware that verifies the JWT token and checks if the session is valid in the database.
func NewWithSession(jwtManager domain.JWTManager, sessionRepo domain.SessionRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token, err := extractToken(c)
			if err != nil {
				return err
			}

			claims, err := jwtManager.VerifyAccessToken(token)
			if err != nil {
				return handleTokenError(err)
			}

			storedToken, err := sessionRepo.GetAccessToken(c.Request().Context(), claims.ID)
			if err != nil || storedToken != token {
				return problem.ErrUnauthorized("invalid session")
			}

			setUser(c, claims)
			return next(c)
		}
	}
}

func extractToken(c *echo.Context) (string, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		return "", problem.ErrUnauthorized("missing Authorization header")
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", problem.ErrUnauthorized("Authorization header must start with 'Bearer '")
	}

	token := parts[1]
	if token == "" {
		return "", problem.ErrUnauthorized("missing token in Authorization header")
	}

	return token, nil
}

func handleTokenError(err error) error {
	if errors.Is(err, domain.ErrTokenExpired) {
		return problem.ErrTokenExpired()
	}
	return problem.ErrUnauthorized()
}

func setUser(c *echo.Context, claims *domain.JWTClaims) {
	c.Set(userKey, claims)
	req := c.Request()
	ctx := context.WithValue(req.Context(), userContextKey, claims)
	c.SetRequest(req.WithContext(ctx))
}

func GetUser(c *echo.Context) *domain.JWTClaims {
	claims, ok := c.Get(userKey).(*domain.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserFromContext(ctx context.Context) *domain.JWTClaims {
	if ctx == nil {
		return nil
	}
	claims, ok := ctx.Value(userContextKey).(*domain.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
