package repository

import (
	"github.com/akfaiz/go-starter-kit/internal/repository/session"
	"github.com/akfaiz/go-starter-kit/internal/repository/user"
	"github.com/akfaiz/go-starter-kit/internal/repository/usertoken"
	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(
		session.NewRepository,
		user.NewRepository,
		usertoken.NewRepository,
	),
)
