package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

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

	cfg := Config{
		App:       loadAppConfig(),
		Auth:      getAuthConfig(),
		Database:  loadDatabaseConfig(),
		Mail:      loadMailConfig(),
		RateLimit: loadRateLimitConfig(),
		Redis:     loadRedisConfig(),
		Server:    loadServerConfig(),
		Telemetry: loadTelemetryConfig(),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	var errs []string

	requiredStringFields := map[string]string{
		"DB_USER":                     c.Database.User,
		"DB_NAME":                     c.Database.Name,
		"JWT_ACCESS_SECRET":           c.Auth.JWT.AccessSecret,
		"JWT_REFRESH_SECRET":          c.Auth.JWT.RefreshSecret,
		"JWT_ACCESS_EXPIRES_IN":       c.Auth.JWT.AccessExpires.String(),
		"JWT_REFRESH_EXPIRES_IN":      c.Auth.JWT.RefreshExpires.String(),
		"MAIL_FROM_ADDRESS":           c.Mail.From.Address,
		"MAIL_FROM_NAME":              c.Mail.From.Name,
		"REDIS_ADDR":                  c.Redis.Addr,
		"REDIS_PREFIX":                c.Redis.Prefix,
		"SERVER_CORS_ORIGINS":         strings.Join(c.Server.CORSOrigins, ","),
		"MAIL_TLS_MODE":               c.Mail.SMTP.TLSMode,
		"OTEL_SERVICE_NAME":           c.Telemetry.ServiceName,
		"OTEL_EXPORTER":               c.Telemetry.Exporter,
		"OTEL_EXPORTER_OTLP_ENDPOINT": c.Telemetry.Endpoint,
	}

	for field, value := range requiredStringFields {
		if strings.TrimSpace(value) == "" {
			errs = append(errs, fmt.Sprintf("%s is required", field))
		}
	}

	if c.Auth.JWT.AccessExpires <= 0 {
		errs = append(errs, "JWT_ACCESS_EXPIRES_IN must be greater than 0")
	}
	if c.Auth.JWT.RefreshExpires <= 0 {
		errs = append(errs, "JWT_REFRESH_EXPIRES_IN must be greater than 0")
	}
	if c.Server.Port <= 0 {
		errs = append(errs, "SERVER_PORT must be greater than 0")
	}
	if c.RateLimit.LoginAttempts <= 0 || c.RateLimit.LoginWindow <= 0 {
		errs = append(errs, "AUTH_LOGIN_LIMIT_ATTEMPTS and AUTH_LOGIN_LIMIT_WINDOW must be greater than 0")
	}
	if c.RateLimit.LoginLockoutThreshold <= 0 || c.RateLimit.LoginLockoutDuration <= 0 {
		errs = append(errs, "AUTH_LOGIN_LOCKOUT_THRESHOLD and AUTH_LOGIN_LOCKOUT_DURATION must be greater than 0")
	}
	if c.RateLimit.RefreshAttemptsPerIP <= 0 || c.RateLimit.RefreshWindow <= 0 {
		errs = append(errs, "AUTH_REFRESH_LIMIT_ATTEMPTS and AUTH_REFRESH_LIMIT_WINDOW must be greater than 0")
	}
	if c.Telemetry.SampleRatio < 0 || c.Telemetry.SampleRatio > 1 {
		errs = append(errs, "OTEL_TRACES_SAMPLER_RATIO must be between 0 and 1")
	}
	if c.Telemetry.ExportTimeout <= 0 {
		errs = append(errs, "OTEL_EXPORT_TIMEOUT must be greater than 0")
	}
	if strings.TrimSpace(c.Telemetry.Endpoint) == "" {
		errs = append(errs, "OTEL_EXPORTER_ENDPOINT is required")
	}
	switch c.Mail.SMTP.TLSMode {
	case "starttls", "tls", "none":
	default:
		errs = append(errs, "MAIL_TLS_MODE must be one of: starttls, tls, none")
	}
	switch c.Telemetry.Exporter {
	case "otlp", "none":
	default:
		errs = append(errs, "OTEL_EXPORTER must be one of: otlp, none")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
