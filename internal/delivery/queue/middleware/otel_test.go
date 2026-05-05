package middleware_test

import (
	"context"
	"errors"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/middleware"
	cerrors "github.com/cockroachdb/errors"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestOtel(t *testing.T) {
	prevPropagator := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTextMapPropagator(prevPropagator)
	})

	otel.SetTextMapPropagator(propagation.TraceContext{})

	t.Run("extracts trace context into the handler context", func(t *testing.T) {
		prevTracerProvider := otel.GetTracerProvider()
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
		otel.SetTracerProvider(tp)
		t.Cleanup(func() {
			_ = tp.Shutdown(context.Background())
			otel.SetTracerProvider(prevTracerProvider)
		})

		wantTraceID := oteltrace.TraceID{
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
			0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		}
		wantSpanID := oteltrace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}

		headers := map[string]string{}
		otel.GetTextMapPropagator().Inject(
			oteltrace.ContextWithSpanContext(context.Background(), oteltrace.NewSpanContext(oteltrace.SpanContextConfig{
				TraceID:    wantTraceID,
				SpanID:     wantSpanID,
				TraceFlags: oteltrace.FlagsSampled,
			})),
			propagation.MapCarrier(headers),
		)

		task := asynq.NewTaskWithHeaders("mail:send", []byte(`{}`), headers)

		handler := asynq.HandlerFunc(func(ctx context.Context, task *asynq.Task) error {
			sc := oteltrace.SpanContextFromContext(ctx)
			require.True(t, sc.IsValid())
			assert.Equal(t, wantTraceID, sc.TraceID())
			assert.Equal(t, "mail:send", task.Type())
			return nil
		})

		err := middleware.Otel(handler).ProcessTask(context.Background(), task)
		require.NoError(t, err)

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Equal(t, "mail:send process", spans[0].Name)
		assert.Equal(t, oteltrace.SpanKindConsumer, spans[0].SpanKind)
		assert.Contains(t, spans[0].Attributes, semconv.MessagingSystemKey.String("asynq"))
		assert.Contains(t, spans[0].Attributes, semconv.MessagingDestinationNameKey.String("mail:send"))
		assert.Contains(t, spans[0].Attributes, semconv.MessagingOperationTypeDeliver)
	})

	t.Run("returns handler error", func(t *testing.T) {
		prevTracerProvider := otel.GetTracerProvider()
		exporter := tracetest.NewInMemoryExporter()
		tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
		otel.SetTracerProvider(tp)
		t.Cleanup(func() {
			_ = tp.Shutdown(context.Background())
			otel.SetTracerProvider(prevTracerProvider)
		})

		wantErr := cerrors.WithStack(errors.New("boom"))
		task := asynq.NewTask("mail:send", []byte(`{}`))

		handler := asynq.HandlerFunc(func(context.Context, *asynq.Task) error {
			return wantErr
		})

		err := middleware.Otel(handler).ProcessTask(context.Background(), task)
		require.ErrorIs(t, err, wantErr)

		spans := exporter.GetSpans()
		require.Len(t, spans, 1)
		assert.Equal(t, codes.Error, spans[0].Status.Code)
		assert.Contains(t, spans[0].Attributes, semconv.ExceptionTypeKey.String("*errors.errorString"))
		assert.Contains(t, spans[0].Attributes, semconv.ExceptionMessageKey.String("boom"))
		assert.NotEmpty(t, findAttr(t, spans[0].Attributes, semconv.ExceptionStacktraceKey).Value.AsString())
	})
}

func findAttr(t *testing.T, attrs []attribute.KeyValue, key attribute.Key) attribute.KeyValue {
	t.Helper()
	for _, attr := range attrs {
		if attr.Key == key {
			return attr
		}
	}
	t.Fatalf("attribute %s not found", key)
	return attribute.KeyValue{}
}
