package telemetry

import (
	stderrors "errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	cerrors "github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// RecordSpanError records error details and stack trace (if available) on a span.
func RecordSpanError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	attrs := []attribute.KeyValue{
		semconv.ExceptionTypeKey.String(ErrorType(err)),
		semconv.ExceptionMessageKey.String(err.Error()),
	}

	if stack, ok := stackTraceString(err); ok {
		attrs = append(attrs, semconv.ExceptionStacktraceKey.String(stack))
	}

	span.SetAttributes(attrs...)
}

// ErrorType returns the first non-wrapper error type in the chain.
func ErrorType(err error) string {
	last := err
	for current := err; current != nil; current = stderrors.Unwrap(current) {
		last = current
		if !isWrapperError(current) {
			return fmt.Sprintf("%T", current)
		}
	}
	return fmt.Sprintf("%T", last)
}

func isWrapperError(err error) bool {
	t := reflect.TypeOf(err)
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	pkg := t.PkgPath()
	switch {
	case strings.HasPrefix(pkg, "github.com/cockroachdb/errors/"):
		return true
	case pkg == "fmt" && t.Name() == "wrapError":
		return true
	case pkg == "errors" && t.Name() == "joinError":
		return true
	}

	return false
}

func stackTraceString(err error) (string, bool) {
	for current := err; current != nil; current = stderrors.Unwrap(current) {
		stack := cerrors.GetReportableStackTrace(current)
		if stack == nil || len(stack.Frames) == 0 {
			continue
		}

		frames := slices.Clone(stack.Frames)
		slices.Reverse(frames)

		var b strings.Builder
		for i, frame := range frames {
			if i > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(formatStackFrame(frame))
		}
		return b.String(), true
	}

	return "", false
}

func formatStackFrame(frame sentry.Frame) string {
	function := firstNonEmpty(frame.Function, frame.Symbol)
	file := firstNonEmpty(frame.Filename, frame.AbsPath)
	line := frame.Lineno

	if function != "" && file != "" && line > 0 {
		return fmt.Sprintf("%s\n\t%s:%d", function, file, line)
	}
	if function != "" && file != "" {
		return fmt.Sprintf("%s\n\t%s", function, file)
	}
	if file != "" && line > 0 {
		return fmt.Sprintf("%s:%d", file, line)
	}
	if function != "" {
		return function
	}

	return fmt.Sprintf("%+v", frame)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
