package middleware

import (
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware/auth"
	"github.com/akfaiz/go-starter-kit/internal/domain"
	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
)

type Middleware struct {
	fx.Out

	Auth echo.MiddlewareFunc `name:"auth"`
}

type Config struct {
	fx.In

	JWTManager  domain.JWTManager
	SessionRepo domain.SessionRepository
}

func New(cfg Config) Middleware {
	return Middleware{
		Auth: auth.NewWithSession(cfg.JWTManager, cfg.SessionRepo),
	}
}
