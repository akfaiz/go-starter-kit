package service

import (
	"github.com/akfaiz/go-starter-kit/internal/service/auth"
	"github.com/akfaiz/go-starter-kit/internal/service/user"
	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(
		auth.NewService,
		user.NewService,
	),
)
