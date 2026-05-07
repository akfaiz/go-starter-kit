package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestFormatRecordMessageAsJSON(t *testing.T) {
	record := slog.NewRecord(time.Unix(0, 0), slog.LevelInfo, "request", 0)
	record.AddAttrs(
		slog.String("method", "GET"),
		slog.Group("error",
			slog.String("msg", "boom now"),
			slog.Int("code", 500),
		),
	)

	inlined := formatRecordMessageAsJSON(record)

	var payload map[string]any
	err := json.Unmarshal([]byte(inlined.Message), &payload)
	require.NoError(t, err)

	assert.Equal(t, "request", payload["msg"])
	assert.Equal(t, "GET", payload["method"])

	errorPayload, ok := payload["error"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "boom now", errorPayload["msg"])
	assert.Equal(t, float64(500), errorPayload["code"])
}
