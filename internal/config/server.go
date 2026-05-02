package config

import (
	"strings"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Server struct {
	Port        int
	CORSOrigins []string
}

func loadServerConfig() Server {
	originsRaw := env.GetString("SERVER_CORS_ORIGINS", "http://localhost:8080")
	origins := make([]string, 0)
	for _, origin := range strings.Split(originsRaw, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		origins = append(origins, trimmed)
	}

	return Server{
		Port:        env.GetInt("SERVER_PORT", 8080),
		CORSOrigins: origins,
	}
}
