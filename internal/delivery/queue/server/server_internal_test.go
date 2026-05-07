package server

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/fx"
)

type captureRecord struct {
	level slog.Level
	msg   string
}

type captureHandler struct {
	mu      sync.Mutex
	records []captureRecord
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	h.records = append(h.records, captureRecord{level: r.Level, msg: r.Message})
	h.mu.Unlock()
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

type fakeLifecycle struct{ hook *fx.Hook }

func (f *fakeLifecycle) Append(h fx.Hook) { f.hook = &h }

func TestRegisterServer_Hooks(t *testing.T) {
	lc := &fakeLifecycle{}
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: "127.0.0.1:0"}, asynq.Config{Concurrency: 1})
	mux := asynq.NewServeMux()
	RegisterServer(lc, srv, mux)

	if lc.hook == nil || lc.hook.OnStart == nil || lc.hook.OnStop == nil {
		t.Fatal("expected lifecycle hooks")
	}

	if err := lc.hook.OnStart(context.Background()); err != nil {
		t.Fatalf("onstart err: %v", err)
	}
	// let goroutine execute run path (will fail with invalid redis addr)
	time.Sleep(10 * time.Millisecond)
	if err := lc.hook.OnStop(context.Background()); err != nil {
		t.Fatalf("onstop err: %v", err)
	}
}

func TestAsynqLogger_MethodsAndBranches(t *testing.T) {
	h := &captureHandler{}
	l := &asynqLogger{l: slog.New(h)}

	l.log(l.l.Info)
	l.log(l.l.Info, 1, "x")
	l.Debug("debug")
	l.Info("info")
	l.Warn("warn")
	l.Error("error")
	l.Fatal("fatal")
	l.Error(errors.New("err"))

	if len(h.records) < 8 {
		t.Fatalf("expected records, got %d", len(h.records))
	}
	if h.records[0].msg != "" {
		t.Fatalf("expected empty msg, got %q", h.records[0].msg)
	}
	if h.records[1].msg != "1x" {
		t.Fatalf("expected fmt fallback msg, got %q", h.records[1].msg)
	}
	if h.records[6].level != slog.LevelError {
		t.Fatalf("expected fatal mapped to error")
	}
}
