package config

import (
	"log"

	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/joho/godotenv"
)

type Config struct {
	App       App
	Auth      Auth
	Database  Database
	Mail      Mail
	RateLimit RateLimit
	Redis     Redis
	Server    Server
	Telemetry Telemetry
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := loadConfig()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	return validator.New().Validate(c)
}

func loadConfig() Config {
	cfg := Config{
		App:       loadAppConfig(),
		Auth:      loadAuthConfig(),
		Database:  loadDatabaseConfig(),
		Mail:      loadMailConfig(),
		RateLimit: loadRateLimitConfig(),
		Redis:     loadRedisConfig(),
		Server:    loadServerConfig(),
		Telemetry: loadTelemetryConfig(),
	}

	return cfg
}
