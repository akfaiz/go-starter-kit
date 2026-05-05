package i18n_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/akfaiz/go-starter-kit/pkg/i18n"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// localeFS returns an in-memory FS with minimal en and id catalogs.
// The YAML is intentionally bare (not pre-wrapped under the locale key) to
// exercise the normalizeLocaleData wrapping path inside LoadWithDefault.
func localeFS() fstest.MapFS {
	return fstest.MapFS{
		"en.yml": &fstest.MapFile{
			Data: []byte("greeting: \"Hello\"\nfarewell: \"Goodbye\"\n"),
		},
		"id.yml": &fstest.MapFile{
			Data: []byte("greeting: \"Halo\"\nfarewell: \"Selamat tinggal\"\n"),
		},
	}
}

// echoCtx builds a minimal *echo.Context whose underlying request carries ctx.
func echoCtx(ctx context.Context) *echo.Context {
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	return echo.New().NewContext(req, rec)
}

// TestLoadWithDefault_Success verifies that a valid FS loads without error.
func TestLoadWithDefault_Success(t *testing.T) {
	err := i18n.LoadWithDefault(localeFS(), "en")
	require.NoError(t, err)
}

// TestLoadWithDefault_InvalidYAML verifies that malformed YAML returns an error.
func TestLoadWithDefault_InvalidYAML(t *testing.T) {
	fs := fstest.MapFS{
		"en.yml": &fstest.MapFile{Data: []byte(":\tinvalid: [yaml")},
	}
	err := i18n.LoadWithDefault(fs, "en")
	assert.Error(t, err)
}

// TestLoadWithDefault_WrapsUnwrappedYAML verifies that bare YAML catalogs (not
// pre-nested under the locale key) are loaded and resolved correctly — this
// exercises the internal normalizeLocaleData wrapping path.
func TestLoadWithDefault_WrapsUnwrappedYAML(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)
	assert.Equal(t, "Hello", i18n.TCtx(ctx, "greeting"))
}

// TestLoadWithDefault_AlreadyWrappedYAML verifies that catalogs already nested
// under the locale key are also loaded and resolved correctly.
func TestLoadWithDefault_AlreadyWrappedYAML(t *testing.T) {
	fs := fstest.MapFS{
		"en.yml": &fstest.MapFile{
			Data: []byte("en:\n  greeting: \"Hello\"\n"),
		},
	}
	require.NoError(t, i18n.LoadWithDefault(fs, "en"))

	ctx, err := i18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)
	assert.Equal(t, "Hello", i18n.TCtx(ctx, "greeting"))
}

// TestWithLocale_KnownLocale verifies that a registered locale is accepted.
func TestWithLocale_KnownLocale(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)
	assert.NotNil(t, ctx)
}

// TestWithLocale_FallbackLocale verifies that the secondary registered locale is accepted.
func TestWithLocale_FallbackLocale(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "id")
	require.NoError(t, err)
	assert.NotNil(t, ctx)
}

// TestWithLocale_UnknownLocale verifies that an unregistered locale falls back to
// the default locale without returning an error (ctxi18n fallback behaviour).
func TestWithLocale_UnknownLocale(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "fr")
	// ctxi18n falls back to the default locale rather than erroring
	require.NoError(t, err)
	assert.NotNil(t, ctx)
	// translation falls back to the English default
	assert.Equal(t, "Hello", i18n.TCtx(ctx, "greeting"))
}

// TestTCtx_ReturnsTranslation verifies that TCtx resolves a key via [context.Context].
func TestTCtx_ReturnsTranslation(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	assert.Equal(t, "Hello", i18n.TCtx(ctx, "greeting"))
}

// TestTCtx_IdLocale verifies that TCtx resolves the correct locale for Indonesian.
func TestTCtx_IdLocale(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "id")
	require.NoError(t, err)

	assert.Equal(t, "Halo", i18n.TCtx(ctx, "greeting"))
}

// TestTCtx_MissingKeyReturnsNonEmpty verifies that an unknown key returns a non-empty fallback.
func TestTCtx_MissingKeyReturnsNonEmpty(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	result := i18n.TCtx(ctx, "nonexistent.key")
	assert.NotEmpty(t, result)
}

// TestT_ReturnsTranslation verifies that T resolves a key via *echo.Context.
func TestT_ReturnsTranslation(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "en")
	require.NoError(t, err)

	assert.Equal(t, "Hello", i18n.T(echoCtx(ctx), "greeting"))
}

// TestT_IdLocale verifies that T uses the locale from the echo context.
func TestT_IdLocale(t *testing.T) {
	require.NoError(t, i18n.LoadWithDefault(localeFS(), "en"))

	ctx, err := i18n.WithLocale(context.Background(), "id")
	require.NoError(t, err)

	assert.Equal(t, "Halo", i18n.T(echoCtx(ctx), "greeting"))
}

// TestT_YAMLExtensions verifies that both .yml and .yaml file extensions are
// handled — exercises the extension-switch path in localeFileFS.ReadFile.
func TestT_YAMLExtensions(t *testing.T) {
	for _, ext := range []string{".yml", ".yaml"} {
		t.Run(ext, func(t *testing.T) {
			fs := fstest.MapFS{
				"en" + ext: &fstest.MapFile{
					Data: []byte("greeting: \"Hello\"\n"),
				},
			}
			require.NoError(t, i18n.LoadWithDefault(fs, "en"))

			ctx, err := i18n.WithLocale(context.Background(), "en")
			require.NoError(t, err)
			assert.Equal(t, "Hello", i18n.TCtx(ctx, "greeting"))
		})
	}
}
