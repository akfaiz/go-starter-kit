package infra

import (
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"go.uber.org/fx"
)

var Module = fx.Module("infra",
	fx.Provide(
		NewDatabase,
		NewRedisClient,
		NewSMTPMailer,
		validator.New,
	),
)
