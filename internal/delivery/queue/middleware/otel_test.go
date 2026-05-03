package middleware_test

import (
	"context"
	"errors"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/middleware"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestOtel(t *testing.T) {
	prevPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTextMapPropagator(prevPropagator)
	})

	otel.SetTextMapPropagator(propagation.TraceContext{})

	t.Run("extracts trace context into the handler context", func(t *testing.T) {
		wantTraceID := trace.TraceID{
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		}
		wantSpanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}

		headers := map[string]string{}
		otel.GetTextMapPropagator().Inject(
			trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
				TraceID:    wantTraceID,
				SpanID:     wantSpanID,
				TraceFlags: trace.FlagsSampled,
			})),
			propagation.MapCarrier(headers),
		)

		task := asynq.NewTaskWithHeaders("mail:send", []byte(`{}`), headers)

		handler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			sc := trace.SpanContextFromContext(ctx)
			require.True(t, sc.IsValid())
			assert.Equal(t, wantTraceID, sc.TraceID())
			assert.Equal(t, "mail:send", task.Type())
			return nil
		})

		err := middleware.Otel(handler).ProcessTask(context.Background(), task)
		require.NoError(t, err)
	})

	t.Run("returns handler error", func(t *testing.T) {
		wantErr := errors.New("boom")
		task := asynq.NewTask("mail:send", []byte(`{}`))

		handler := asynq.HandlerFunc(func(context.Context, *asynq.Task) error {
			return wantErr
		})

		err := middleware.Otel(handler).ProcessTask(context.Background(), task)
		require.ErrorIs(t, err, wantErr)
	})
}
