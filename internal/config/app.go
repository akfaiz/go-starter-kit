package config

import "github.com/akfaiz/go-starter-kit/pkg/env"

type App struct {
	Name            string
	FrontendBaseURL string
	LogLevel        string
	LogFormat       string
}

func loadAppConfig() App {
	return App{
		Name:            env.GetString("APP_NAME", "go-starter-kit"),
		FrontendBaseURL: env.GetString("FRONTEND_BASE_URL", "http://localhost:8080"),
		LogLevel:        env.GetString("LOG_LEVEL", "debug"),
		LogFormat:       env.GetString("LOG_FORMAT", "json"),
	}
}
