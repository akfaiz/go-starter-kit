package config

import (
	"time"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Auth struct {
	ResetPasswordExpiration time.Duration
	VerificationExpiration  time.Duration
	JWT                     JWT
}

type JWT struct {
	AccessSecret   string        `validate:"required" label:"JWT_ACCESS_SECRET"`
	RefreshSecret  string        `validate:"required" label:"JWT_REFRESH_SECRET"`
	AccessExpires  time.Duration `validate:"gt=0"     label:"JWT_ACCESS_EXPIRES_IN"`
	RefreshExpires time.Duration `validate:"gt=0"     label:"JWT_REFRESH_EXPIRES_IN"`
}

func loadAuthConfig() Auth {
	return Auth{
		ResetPasswordExpiration: 60 * time.Minute,
		VerificationExpiration:  60 * time.Minute,
		JWT: JWT{
			AccessSecret:   env.GetString("JWT_ACCESS_SECRET"),
			RefreshSecret:  env.GetString("JWT_REFRESH_SECRET"),
			AccessExpires:  env.GetDuration("JWT_ACCESS_EXPIRES_IN"),
			RefreshExpires: env.GetDuration("JWT_REFRESH_EXPIRES_IN"),
		},
	}
}
