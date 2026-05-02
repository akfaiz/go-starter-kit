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
	AccessSecret   string
	RefreshSecret  string
	AccessExpires  time.Duration
	RefreshExpires time.Duration
}

func getAuthConfig() Auth {
	return Auth{
		ResetPasswordExpiration: 60 * time.Minute,
		VerificationExpiration:  60 * time.Minute,
		JWT: JWT{
			AccessSecret:   env.MustGetString("JWT_ACCESS_SECRET"),
			RefreshSecret:  env.MustGetString("JWT_REFRESH_SECRET"),
			AccessExpires:  env.MustGetDuration("JWT_ACCESS_EXPIRES_IN"),
			RefreshExpires: env.MustGetDuration("JWT_REFRESH_EXPIRES_IN"),
		},
	}
}
