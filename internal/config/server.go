package config

import (
	"strings"
	"time"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Server struct {
	Port              int           `validate:"gt=0"           label:"SERVER_PORT"`
	CORSOrigins       []string      `validate:"required,min=1" label:"SERVER_CORS_ORIGINS"`
	ReadHeaderTimeout time.Duration `validate:"gt=0"           label:"SERVER_READ_HEADER_TIMEOUT"`
	RateLimitEnabled  bool          `                          label:"SERVER_RATE_LIMIT_ENABLED"`
	RateLimitRequests float64       `                          label:"SERVER_RATE_LIMIT_REQUESTS"`
	RateLimitBurst    int           `                          label:"SERVER_RATE_LIMIT_BURST"`
}

func loadServerConfig() Server {
	originsRaw := env.GetString("SERVER_CORS_ORIGINS", "http://localhost:8080")
	origins := make([]string, 0)
	for origin := range strings.SplitSeq(originsRaw, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		origins = append(origins, trimmed)
	}

	return Server{
		Port:              env.GetInt("SERVER_PORT", 8080),
		CORSOrigins:       origins,
		ReadHeaderTimeout: env.GetDuration("SERVER_READ_HEADER_TIMEOUT", 10*time.Second),
		RateLimitEnabled:  env.GetBool("SERVER_RATE_LIMIT_ENABLED", true),
		RateLimitRequests: env.GetFloat("SERVER_RATE_LIMIT_REQUESTS", 20),
		RateLimitBurst:    env.GetInt("SERVER_RATE_LIMIT_BURST", 40),
	}
}
