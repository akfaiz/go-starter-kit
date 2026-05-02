package bunslog

import (
	"context"
	"database/sql"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Option func(*QueryHook)

func WithLogger(logger *slog.Logger) Option {
	return func(qh *QueryHook) {
		qh.logger = logger
	}
}

func WithEnabled(enabled bool) Option {
	return func(qh *QueryHook) {
		qh.enabled = enabled
	}
}

func WithVerbose(verbose bool) Option {
	return func(qh *QueryHook) {
		qh.verbose = verbose
	}
}

type QueryHook struct {
	logger  *slog.Logger
	enabled bool
	verbose bool
}

var _ bun.QueryHook = (*QueryHook)(nil)

func NewQueryHook(opts ...Option) *QueryHook {
	qh := &QueryHook{
		logger:  slog.Default(),
		enabled: true,
		verbose: false,
	}
	for _, opt := range opts {
		opt(qh)
	}
	return qh
}

func (h *QueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !h.enabled {
		return
	}

	if !h.verbose {
		switch event.Err {
		case nil, sql.ErrNoRows, sql.ErrTxDone:
			return
		}
	}

	level := slog.LevelInfo
	latency := time.Since(event.StartTime)

	attrs := []slog.Attr{
		slog.String("operation", event.Operation()),
		slog.String("query", h.formatQuery(event.Query)),
		slog.Float64("latency", float64(latency.Microseconds())/1000), // in ms
	}
	if event.Err != nil {
		level = slog.LevelError
		attrs = append(attrs, slog.GroupAttrs("error",
			slog.String("msg", event.Err.Error()),
			slog.String("type", reflect.TypeOf(event.Err).String()),
		))
	}

	h.logger.LogAttrs(ctx, level, "bun.query", attrs...)
}

func (b *QueryHook) formatQuery(query string) string {
	q := strings.ReplaceAll(query, `"`, "")
	return strings.TrimSpace(q)
}
