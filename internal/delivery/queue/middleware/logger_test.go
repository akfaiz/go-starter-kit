package middleware_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/middleware"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("logs successful processing with messaging fields", func(t *testing.T) {
		var buf bytes.Buffer
		prev := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{})))
		t.Cleanup(func() {
			slog.SetDefault(prev)
		})

		handler := asynq.HandlerFunc(func(context.Context, *asynq.Task) error {
			return nil
		})

		err := middleware.Logger(handler).ProcessTask(context.Background(), asynq.NewTask("mail:send", []byte(`{}`)))
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "task processing started")
		assert.Contains(t, out, "task processed successfully")
		assert.Contains(t, out, "messaging.system=asynq")
		assert.Contains(t, out, "messaging.destination.name=mail:send")
		assert.Contains(t, out, "messaging.operation.type=process")
		assert.Contains(t, out, "latency=")
	})

	t.Run("logs failures with error and messaging fields", func(t *testing.T) {
		var buf bytes.Buffer
		prev := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{})))
		t.Cleanup(func() {
			slog.SetDefault(prev)
		})

		wantErr := errors.New("boom")
		handler := asynq.HandlerFunc(func(context.Context, *asynq.Task) error {
			return wantErr
		})

		err := middleware.Logger(handler).ProcessTask(context.Background(), asynq.NewTask("mail:send", []byte(`{}`)))
		require.ErrorIs(t, err, wantErr)

		out := buf.String()
		assert.Contains(t, out, "task processing started")
		assert.Contains(t, out, "task processing failed")
		assert.Contains(t, out, "messaging.system=asynq")
		assert.Contains(t, out, "messaging.destination.name=mail:send")
		assert.Contains(t, out, "messaging.operation.type=process")
		assert.Contains(t, out, "error=")
	})
}
