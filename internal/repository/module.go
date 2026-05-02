package repository

import (
	"github.com/akfaiz/go-starter-kit/internal/repository/user"
	"github.com/akfaiz/go-starter-kit/internal/repository/usertoken"
	"go.uber.org/fx"
)

var Module = fx.Module("repository",
	fx.Provide(
		user.NewRepository,
		usertoken.NewRepository,
	),
)
