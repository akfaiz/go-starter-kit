package telemetry_test

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestTelemetry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Telemetry Suite")
}

var _ = Describe("Tracing", Label("unit"), func() {
	var (
		exporter *tracetest.InMemoryExporter
		tp       *trace.TracerProvider
	)

	BeforeEach(func() {
		exporter = tracetest.NewInMemoryExporter()
		tp = trace.NewTracerProvider(trace.WithBatcher(exporter))
	})

	Describe("StartSpan", func() {
		It("should use the caller's function name as the span name", func() {
			tracer := tp.Tracer("test")
			ctx := context.Background()

			// Call a helper function to simulate a service/repository call
			ctx, span := callMe(ctx, tracer)
			span.End()

			err := tp.ForceFlush(ctx)
			Expect(err).NotTo(HaveOccurred())

			spans := exporter.GetSpans()
			Expect(spans).To(HaveLen(1))
			Expect(spans[0].Name).To(Equal("telemetry_test.callMe"))
		})

		It("should handle anonymous functions or unknown callers gracefully", func() {
			tracer := tp.Tracer("test")
			ctx := context.Background()

			func() {
				var span trace.ReadOnlySpan
				_, s := telemetry.StartSpan(ctx, tracer)
				span = s.(trace.ReadOnlySpan)
				defer s.End()

				Expect(span.Name()).To(ContainSubstring("func1"))
			}()
		})
	})
})

func callMe(ctx context.Context, tracer oteltrace.Tracer) (context.Context, oteltrace.Span) {
	return telemetry.StartSpan(ctx, tracer)
}
