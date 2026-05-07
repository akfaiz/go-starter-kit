package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConsoleHandler_DefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newConsoleHandler("invalid", &buf))

	logger.Info("hello", slog.String("key", "value"))

	assert.Contains(t, buf.String(), `"msg":"hello"`)
	assert.Contains(t, buf.String(), `"key":"value"`)
}

func TestNewConsoleHandler_UsesTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newConsoleHandler("text", &buf))

	logger.Info("hello", slog.String("key", "value"))

	assert.Contains(t, buf.String(), "msg=hello")
	assert.Contains(t, buf.String(), "key=value")
}

func TestResolveLogLevel(t *testing.T) {
	assert.Equal(t, slog.LevelDebug, resolveLogLevel("debug"))
	assert.Equal(t, slog.LevelInfo, resolveLogLevel("info"))
	assert.Equal(t, slog.LevelWarn, resolveLogLevel("warning"))
	assert.Equal(t, slog.LevelWarn, resolveLogLevel("warn"))
	assert.Equal(t, slog.LevelError, resolveLogLevel("error"))
	assert.Equal(t, disabledLogLevel, resolveLogLevel("disabled"))
	assert.Equal(t, disabledLogLevel, resolveLogLevel("off"))
	assert.Equal(t, slog.LevelInfo, resolveLogLevel("unknown"))
}
