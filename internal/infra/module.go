package infra

import "go.uber.org/fx"

var Module = fx.Module("provider",
	fx.Provide(
		NewDatabase,
		NewRedisClient,
		NewSMTPMailer,
	),
)
