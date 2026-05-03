package repository

import (
	"github.com/akfaiz/go-starter-kit/internal/repository/passwordresettoken"
	"github.com/akfaiz/go-starter-kit/internal/repository/session"
	"github.com/akfaiz/go-starter-kit/internal/repository/user"
	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(
		session.NewRepository,
		user.NewRepository,
		passwordresettoken.NewRepository,
	),
)
