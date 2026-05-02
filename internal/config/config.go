package config

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App      App
	Auth     Auth
	Database Database
	Mail     Mail
	Server   Server
}

func Load() (Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg := Config{
		App:      loadAppConfig(),
		Auth:     getAuthConfig(),
		Database: loadDatabaseConfig(),
		Mail:     loadMailConfig(),
		Server:   loadServerConfig(),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	var errs []string

	requiredStringFields := map[string]string{
		"DB_USER":                c.Database.User,
		"DB_NAME":                c.Database.Name,
		"JWT_ACCESS_SECRET":      c.Auth.JWT.AccessSecret,
		"JWT_REFRESH_SECRET":     c.Auth.JWT.RefreshSecret,
		"JWT_ACCESS_EXPIRES_IN":  c.Auth.JWT.AccessExpires.String(),
		"JWT_REFRESH_EXPIRES_IN": c.Auth.JWT.RefreshExpires.String(),
		"MAIL_FROM_ADDRESS":      c.Mail.From.Address,
		"MAIL_FROM_NAME":         c.Mail.From.Name,
		"SERVER_CORS_ORIGINS":    strings.Join(c.Server.CORSOrigins, ","),
		"MAIL_TLS_MODE":          c.Mail.SMTP.TLSMode,
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
	switch c.Mail.SMTP.TLSMode {
	case "starttls", "tls", "none":
	default:
		errs = append(errs, "MAIL_TLS_MODE must be one of: starttls, tls, none")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
