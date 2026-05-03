package env_test

import (
	"os"
	"testing"
	"time"

	"github.com/akfaiz/go-starter-kit/pkg/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAndLookupHelpers(t *testing.T) {
	t.Setenv("ENV_TEST_INT", " 42 ")
	t.Setenv("ENV_TEST_BOOL", " true ")
	t.Setenv("ENV_TEST_FLOAT", " 3.14 ")
	t.Setenv("ENV_TEST_DURATION", " 5s ")

	i, ok, err := env.LookupInt("ENV_TEST_INT")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 42, i)

	b, ok, err := env.LookupBool("ENV_TEST_BOOL")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, b)

	f, ok, err := env.LookupFloat("ENV_TEST_FLOAT")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3.14, f)

	d, ok, err := env.LookupDuration("ENV_TEST_DURATION")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 5*time.Second, d)

	assert.Equal(t, "fallback", env.GetString("ENV_TEST_MISSING", "fallback"))
	assert.Equal(t, 7, env.GetInt("ENV_TEST_MISSING_INT", 7))
	assert.True(t, env.GetBool("ENV_TEST_MISSING_BOOL", true))
	assert.Equal(t, 2.5, env.GetFloat("ENV_TEST_MISSING_FLOAT", 2.5))
	assert.Equal(t, 9*time.Second, env.GetDuration("ENV_TEST_MISSING_DURATION", 9*time.Second))
}

func TestInvalidLookupReturnsError(t *testing.T) {
	t.Setenv("ENV_TEST_BAD_INT", "abc")
	t.Setenv("ENV_TEST_BAD_BOOL", "not-bool")
	t.Setenv("ENV_TEST_BAD_FLOAT", "nanx")
	t.Setenv("ENV_TEST_BAD_DURATION", "1 second")

	_, ok, err := env.LookupInt("ENV_TEST_BAD_INT")
	assert.True(t, ok)
	assert.Error(t, err)
	_, ok, err = env.LookupBool("ENV_TEST_BAD_BOOL")
	assert.True(t, ok)
	assert.Error(t, err)
	_, ok, err = env.LookupFloat("ENV_TEST_BAD_FLOAT")
	assert.True(t, ok)
	assert.Error(t, err)
	_, ok, err = env.LookupDuration("ENV_TEST_BAD_DURATION")
	assert.True(t, ok)
	assert.Error(t, err)
}

func TestMain(m *testing.M) {
	// Ensure stale env from shell does not affect tests.
	_ = os.Unsetenv("ENV_TEST_MISSING")
	os.Exit(m.Run())
}
