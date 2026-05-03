package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAndLookupHelpers(t *testing.T) {
	t.Setenv("ENV_TEST_INT", " 42 ")
	t.Setenv("ENV_TEST_BOOL", " true ")
	t.Setenv("ENV_TEST_FLOAT", " 3.14 ")
	t.Setenv("ENV_TEST_DURATION", " 5s ")

	i, ok, err := LookupInt("ENV_TEST_INT")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 42, i)

	b, ok, err := LookupBool("ENV_TEST_BOOL")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, b)

	f, ok, err := LookupFloat("ENV_TEST_FLOAT")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3.14, f)

	d, ok, err := LookupDuration("ENV_TEST_DURATION")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 5*time.Second, d)

	assert.Equal(t, "fallback", GetString("ENV_TEST_MISSING", "fallback"))
	assert.Equal(t, 7, GetInt("ENV_TEST_MISSING_INT", 7))
	assert.True(t, GetBool("ENV_TEST_MISSING_BOOL", true))
	assert.Equal(t, 2.5, GetFloat("ENV_TEST_MISSING_FLOAT", 2.5))
	assert.Equal(t, 9*time.Second, GetDuration("ENV_TEST_MISSING_DURATION", 9*time.Second))
}

func TestInvalidLookupReturnsError(t *testing.T) {
	t.Setenv("ENV_TEST_BAD_INT", "abc")
	t.Setenv("ENV_TEST_BAD_BOOL", "not-bool")
	t.Setenv("ENV_TEST_BAD_FLOAT", "nanx")
	t.Setenv("ENV_TEST_BAD_DURATION", "1 second")

	_, ok, err := LookupInt("ENV_TEST_BAD_INT")
	assert.True(t, ok)
	assert.Error(t, err)
	_, ok, err = LookupBool("ENV_TEST_BAD_BOOL")
	assert.True(t, ok)
	assert.Error(t, err)
	_, ok, err = LookupFloat("ENV_TEST_BAD_FLOAT")
	assert.True(t, ok)
	assert.Error(t, err)
	_, ok, err = LookupDuration("ENV_TEST_BAD_DURATION")
	assert.True(t, ok)
	assert.Error(t, err)
}

func TestMustGetPanicsOnMissingAndInvalid(t *testing.T) {
	assert.Panics(t, func() { MustGetString("ENV_TEST_MUST_MISSING_STR") })
	assert.Panics(t, func() { MustGetInt("ENV_TEST_MUST_MISSING_INT") })
	assert.Panics(t, func() { MustGetBool("ENV_TEST_MUST_MISSING_BOOL") })
	assert.Panics(t, func() { MustGetFloat("ENV_TEST_MUST_MISSING_FLOAT") })
	assert.Panics(t, func() { MustGetDuration("ENV_TEST_MUST_MISSING_DURATION") })

	t.Setenv("ENV_TEST_MUST_BAD_INT", "abc")
	t.Setenv("ENV_TEST_MUST_BAD_BOOL", "abc")
	t.Setenv("ENV_TEST_MUST_BAD_FLOAT", "abc")
	t.Setenv("ENV_TEST_MUST_BAD_DURATION", "abc")

	assert.Panics(t, func() { MustGetInt("ENV_TEST_MUST_BAD_INT") })
	assert.Panics(t, func() { MustGetBool("ENV_TEST_MUST_BAD_BOOL") })
	assert.Panics(t, func() { MustGetFloat("ENV_TEST_MUST_BAD_FLOAT") })
	assert.Panics(t, func() { MustGetDuration("ENV_TEST_MUST_BAD_DURATION") })
}

func TestMustGetReturnsParsedValues(t *testing.T) {
	t.Setenv("ENV_TEST_MUST_STR", "hello")
	t.Setenv("ENV_TEST_MUST_INT", "12")
	t.Setenv("ENV_TEST_MUST_BOOL", "true")
	t.Setenv("ENV_TEST_MUST_FLOAT", "1.5")
	t.Setenv("ENV_TEST_MUST_DURATION", "3m")

	assert.Equal(t, "hello", MustGetString("ENV_TEST_MUST_STR"))
	assert.Equal(t, 12, MustGetInt("ENV_TEST_MUST_INT"))
	assert.True(t, MustGetBool("ENV_TEST_MUST_BOOL"))
	assert.Equal(t, 1.5, MustGetFloat("ENV_TEST_MUST_FLOAT"))
	assert.Equal(t, 3*time.Minute, MustGetDuration("ENV_TEST_MUST_DURATION"))
}

func TestMain(m *testing.M) {
	// Ensure stale env from shell does not affect tests.
	_ = os.Unsetenv("ENV_TEST_MISSING")
	os.Exit(m.Run())
}
