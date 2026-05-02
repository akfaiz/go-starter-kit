package provider

import "go.uber.org/fx"

var Module = fx.Module("provider",
	fx.Provide(
		NewSMTPMailer,
	),
)
