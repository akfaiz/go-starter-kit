package telemetry

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	cerrors "github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackTraceString_UsesReadableFormat(t *testing.T) {
	stack, ok := stackTraceString(cerrors.WithStack(errors.New("boom")))
	require.True(t, ok)

	assert.Contains(t, stack, "\n\t")
	assert.False(t, strings.Contains(stack, "{Function:"))
}

func TestExceptionType_PrefersFirstNonWrapper(t *testing.T) {
	err := cerrors.WithStack(fmt.Errorf("context: %w", io.EOF))
	assert.Equal(t, "*errors.errorString", ErrorType(err))
}
