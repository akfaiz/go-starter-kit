package i18n

import (
	"bytes"
	"context"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/invopop/ctxi18n"
	"github.com/invopop/ctxi18n/i18n"
	"github.com/invopop/yaml"
	"github.com/labstack/echo/v5"
)

// T is a helper function to translate messages using the echo context.
func T(c *echo.Context, key string, args ...any) string {
	return i18n.T(c.Request().Context(), key, args...)
}

// TCtx is a helper function to translate messages using a [context.Context].
func TCtx(ctx context.Context, key string, args ...any) string {
	return i18n.T(ctx, key, args...)
}

// LoadWithDefault performs the regular load operation, but will merge
// the default locale with every other locale, ensuring that every text
// has at least the value from the default locale.
func LoadWithDefault(fs fs.FS, locale i18n.Code) error {
	return ctxi18n.LoadWithDefault(localeFileFS{FS: fs}, locale)
}

// WithLocale tries to match the provided code with a locale and ensures.
func WithLocale(ctx context.Context, locale string) (context.Context, error) {
	return ctxi18n.WithLocale(ctx, locale)
}

type localeFileFS struct {
	fs.FS
}

func (l localeFileFS) ReadFile(name string) ([]byte, error) {
	data, err := fs.ReadFile(l.FS, name)
	if err != nil {
		return nil, err
	}

	switch filepath.Ext(name) {
	case ".yml", ".yaml", ".json":
		return normalizeLocaleData(filepath.Base(name), data)
	default:
		return data, nil
	}
}

func normalizeLocaleData(filename string, data []byte) ([]byte, error) {
	locale := strings.TrimSuffix(filename, filepath.Ext(filename))

	var root any
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	if m, ok := root.(map[string]any); ok {
		if _, exists := m[locale]; exists {
			return data, nil
		}
	}

	wrapped := map[string]any{
		locale: root,
	}

	b, err := yaml.Marshal(wrapped)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(b), nil
}
