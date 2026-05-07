package server_test

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/server"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/fx"
)

func TestNewQueueMetricsCollector_Disabled(t *testing.T) {
	collector, err := server.NewQueueMetricsCollector(
		config.Config{Queue: config.Queue{MetricsEnabled: false}},
		noop.NewMeterProvider(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if collector == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestNewQueueMetricsCollector_Enabled(t *testing.T) {
	collector, err := server.NewQueueMetricsCollector(
		config.Config{
			Queue: config.Queue{MetricsEnabled: true},
			Redis: config.Redis{Addr: "127.0.0.1:6379", DB: 0},
		},
		noop.NewMeterProvider(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if collector == nil {
		t.Fatal("expected non-nil collector")
	}
}

func TestRegisterQueueMetricsCollector_NoOpForNilOrDisabled(t *testing.T) {
	lc := &fakeLifecycle{}

	server.RegisterQueueMetricsCollector(lc, nil)
	if lc.hooks != 0 {
		t.Fatalf("expected no hooks for nil collector, got %d", lc.hooks)
	}

	disabledCollector, err := server.NewQueueMetricsCollector(
		config.Config{Queue: config.Queue{MetricsEnabled: false}},
		noop.NewMeterProvider(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	server.RegisterQueueMetricsCollector(lc, disabledCollector)
	if lc.hooks != 0 {
		t.Fatalf("expected no hooks for disabled collector, got %d", lc.hooks)
	}
}

type fakeLifecycle struct{ hooks int }

func (f *fakeLifecycle) Append(h fx.Hook) {
	_ = h
	f.hooks++
}

func (f *fakeLifecycle) Start(context.Context) error { return nil }
func (f *fakeLifecycle) Stop(context.Context) error  { return nil }
