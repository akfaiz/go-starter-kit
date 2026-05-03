package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() Config {
	return Config{
		App: App{Name: "app"},
		Auth: Auth{JWT: JWT{
			AccessSecret:   "access-secret",
			RefreshSecret:  "refresh-secret",
			AccessExpires:  15 * time.Minute,
			RefreshExpires: 24 * time.Hour,
		}},
		Database: Database{User: "dbuser", Name: "dbname"},
		Mail: Mail{
			SMTP: MailSMTP{TLSMode: "starttls"},
			From: MailFrom{Address: "noreply@example.com", Name: "Example"},
		},
		RateLimit: RateLimit{
			LoginAttempts:         5,
			LoginWindow:           10 * time.Minute,
			LoginLockoutThreshold: 5,
			LoginLockoutDuration:  15 * time.Minute,
			RefreshAttemptsPerIP:  20,
			RefreshWindow:         10 * time.Minute,
		},
		Redis:  Redis{Addr: "localhost:6379", Prefix: "gsk"},
		Server: Server{Port: 8080, CORSOrigins: []string{"http://localhost:8080"}},
		Telemetry: Telemetry{
			ServiceName:   "go-starter-kit",
			Exporter:      "otlp",
			Endpoint:      "jaeger:4317",
			SampleRatio:   1,
			ExportTimeout: 5 * time.Second,
		},
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := validConfig()
	assert.NoError(t, cfg.validate())
}

func TestValidate_FailureAggregatesErrors(t *testing.T) {
	cfg := validConfig()
	cfg.Database.User = ""
	cfg.Auth.JWT.AccessExpires = 0
	cfg.Mail.SMTP.TLSMode = "invalid"
	cfg.Telemetry.Exporter = "zipkin"
	cfg.Telemetry.SampleRatio = 2

	err := cfg.validate()
	require.Error(t, err)

	msg := err.Error()
	checks := []string{
		"DB_USER is required",
		"JWT_ACCESS_EXPIRES_IN must be greater than 0",
		"MAIL_TLS_MODE must be one of: starttls, tls, none",
		"OTEL_EXPORTER must be one of: otlp, none",
		"OTEL_TRACES_SAMPLER_RATIO must be between 0 and 1",
	}
	for _, want := range checks {
		assert.True(t, strings.Contains(msg, want), "validate() error %q does not contain %q", msg, want)
	}
}
