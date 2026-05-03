package handler

import "go.uber.org/fx"

var Module = fx.Module("handler",
	fx.Provide(
		NewHealthCheckHandler,
		NewAuthHandler,
		NewProfileHandler,
		NewUserHandler,
	),
)
