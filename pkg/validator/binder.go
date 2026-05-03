package validator

import (
	"github.com/labstack/echo/v5"
)

// Binder wraps the default Echo binder and adds automatic validation.
type Binder struct {
	defaultBinder echo.Binder
	validator     *Validate
}

var _ echo.Binder = (*Binder)(nil)

// NewBinder creates a new Binder instance with the provided default binder and validator.
func NewBinder(defaultBinder echo.Binder, v *Validate) *Binder {
	return &Binder{defaultBinder: defaultBinder, validator: v}
}

// Bind binds the request data to the provided interface and then validates it.
//
// It leverages Echo's default binder for data extraction and the validator
// for validation with the request context to support i18n.
func (b *Binder) Bind(c *echo.Context, i any) error {
	// 1. Bind the data using the default binder
	if err := b.defaultBinder.Bind(c, i); err != nil {
		return err
	}

	// 2. Automatically validate the bound struct with context (for i18n)
	return b.validator.ValidateContext(c.Request().Context(), i)
}
