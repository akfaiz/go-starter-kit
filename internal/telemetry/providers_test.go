package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/telemetry"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
)

type hookLifecycle struct {
	hooks []fx.Hook
}

func (l *hookLifecycle) Append(h fx.Hook) {
	l.hooks = append(l.hooks, h)
}

func TestNewTracerProvider_Disabled(t *testing.T) {
	tp, err := telemetry.NewTracerProvider(config.Config{Telemetry: config.Telemetry{Enabled: false}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp == nil {
		t.Fatal("expected tracer provider")
	}

	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "span")
	if span.IsRecording() {
		t.Fatal("expected non-recording span for disabled telemetry")
	}
	span.End()
}

func TestNewTracerProvider_ExporterNone(t *testing.T) {
	tp, err := telemetry.NewTracerProvider(config.Config{Telemetry: config.Telemetry{Enabled: true, Exporter: "none"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tp == nil {
		t.Fatal("expected tracer provider")
	}
}

func TestNewMeterProvider_Disabled(t *testing.T) {
	mp, err := telemetry.NewMeterProvider(config.Config{Telemetry: config.Telemetry{Enabled: false}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mp == nil {
		t.Fatal("expected meter provider")
	}
}

func TestNewMeterProvider_ExporterNone(t *testing.T) {
	mp, err := telemetry.NewMeterProvider(config.Config{Telemetry: config.Telemetry{Enabled: true, Exporter: "none"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mp == nil {
		t.Fatal("expected meter provider")
	}
}

func TestNewLoggerProvider_Disabled(t *testing.T) {
	lp, err := telemetry.NewLoggerProvider(config.Config{Telemetry: config.Telemetry{Enabled: false}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lp == nil {
		t.Fatal("expected logger provider")
	}
}

func TestNewLoggerProvider_ExporterNone(t *testing.T) {
	lp, err := telemetry.NewLoggerProvider(config.Config{Telemetry: config.Telemetry{Enabled: true, Exporter: "none"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lp == nil {
		t.Fatal("expected logger provider")
	}
}

func TestRegisterLifecycle_AppendsHookAndStopsProviders(t *testing.T) {
	lc := &hookLifecycle{}
	cfg := config.Config{Telemetry: config.Telemetry{ExportTimeout: time.Second}}

	tp := sdktrace.NewTracerProvider()
	mp := sdkmetric.NewMeterProvider()
	lp := sdklog.NewLoggerProvider()

	telemetry.RegisterLifecycle(lc, cfg, tp, mp, lp)

	if len(lc.hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(lc.hooks))
	}

	onStop := lc.hooks[0].OnStop
	if onStop == nil {
		t.Fatal("expected OnStop hook")
	}
	if err := onStop(context.Background()); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}
}

func TestNewTracerProvider_EnabledBranch(t *testing.T) {
	tp, err := telemetry.NewTracerProvider(config.Config{Telemetry: config.Telemetry{
		Enabled:       true,
		Exporter:      "otlp",
		Endpoint:      "bad endpoint",
		Insecure:      true,
		ExportTimeout: time.Millisecond,
		SampleRatio:   1,
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = tp.Shutdown(context.Background())
}

func TestNewMeterProvider_EnabledBranch(t *testing.T) {
	mp, err := telemetry.NewMeterProvider(config.Config{Telemetry: config.Telemetry{
		Enabled:       true,
		Exporter:      "otlp",
		Endpoint:      "bad endpoint",
		Insecure:      true,
		ExportTimeout: time.Millisecond,
		SampleRatio:   1,
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = mp.Shutdown(context.Background())
}

func TestNewLoggerProvider_EnabledBranch(t *testing.T) {
	lp, err := telemetry.NewLoggerProvider(config.Config{Telemetry: config.Telemetry{
		Enabled:       true,
		Exporter:      "otlp",
		Endpoint:      "bad endpoint",
		Insecure:      true,
		ExportTimeout: time.Millisecond,
		SampleRatio:   1,
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = lp.Shutdown(context.Background())
}
