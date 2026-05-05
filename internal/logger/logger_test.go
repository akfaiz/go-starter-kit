package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	cerrors "github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
)

func TestCustomHandler_LogsErrorStackWhenAvailable(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(&customHandler{
		Handler: slog.NewTextHandler(&buf, &slog.HandlerOptions{}),
	})

	logger.ErrorContext(context.Background(), "failed", slog.Any("error", cerrors.WithStack(errors.New("boom"))))

	assert.Contains(t, buf.String(), "error_type=*errors.errorString")
	assert.Contains(t, buf.String(), "error_stack=")
	assert.Contains(t, buf.String(), "File:")
	assert.NotContains(t, buf.String(), "abs_path")
}
