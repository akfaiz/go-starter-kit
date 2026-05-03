package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setValidEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_USER", "dbuser")
	t.Setenv("DB_NAME", "dbname")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret")
	t.Setenv("JWT_ACCESS_EXPIRES_IN", "15m")
	t.Setenv("JWT_REFRESH_EXPIRES_IN", "24h")
	t.Setenv("MAIL_FROM_ADDRESS", "noreply@example.com")
	t.Setenv("MAIL_FROM_NAME", "Example")
	t.Setenv("MAIL_TLS_MODE", "starttls")
	t.Setenv("OTEL_EXPORTER", "otlp")
	t.Setenv("OTEL_TRACES_SAMPLER_RATIO", "1")
	t.Setenv("OTEL_EXPORT_TIMEOUT", "5s")
}

func TestLoadConfig_Success(t *testing.T) {
	setValidEnv(t)

	cfg := loadConfig()
	err := cfg.Validate()
	require.NoError(t, err)
	assert.Equal(t, "dbuser", cfg.Database.User)
	assert.Equal(t, "dbname", cfg.Database.Name)
}

func TestLoadDatabaseConfig_ReturnsErrorOnMissingUser(t *testing.T) {
	setValidEnv(t)
	t.Setenv("DB_USER", "")

	cfg := loadConfig()
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_USER is a required field")
}

func TestLoadConfig_ReturnsErrorOnInvalidMailTLSMode(t *testing.T) {
	setValidEnv(t)
	t.Setenv("MAIL_TLS_MODE", "invalid")

	cfg := loadConfig()
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MAIL_TLS_MODE must be one of [starttls tls none]")
}
