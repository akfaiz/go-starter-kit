package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/labstack/echo/v5"
)

const typeLabelNumber = "number"

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
		if unmarshalTypeErr := extractUnmarshalTypeError(err); unmarshalTypeErr != nil {
			field := normalizeJSONField(unmarshalTypeErr.Field)
			message := fmt.Sprintf("%s must be a %s", prettyFieldName(field), unmarshalTypeErr.Type.String())

			return problem.ErrBadRequest().
				WithDetail("One or more fields have invalid formats.").
				WithErrors(NewErrors(FieldError{Field: field, Message: message})).
				WithCause(unmarshalTypeErr)
		}

		if bindingTypeErrors := collectBindingTypeErrors(c, i); bindingTypeErrors != nil {
			return problem.ErrBadRequest().
				WithDetail("One or more fields have invalid formats.").
				WithErrors(bindingTypeErrors).
				WithCause(err)
		}

		originalErr := errors.Unwrap(err)
		if originalErr == nil {
			originalErr = err
		}
		return problem.ErrBadRequest().WithCause(originalErr).WithDetail(originalErr.Error())
	}

	if bindingTypeErrors := collectBindingTypeErrors(c, i); bindingTypeErrors != nil {
		return problem.ErrBadRequest().
			WithDetail("One or more fields have invalid formats.").
			WithErrors(bindingTypeErrors)
	}

	// 2. Automatically validate the bound struct with context (for i18n)
	return b.validator.ValidateContext(c.Request().Context(), i)
}

func collectBindingTypeErrors(c *echo.Context, i any) *ValidationError {
	t := reflect.TypeOf(i)
	if t == nil {
		return nil
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	errList := make([]FieldError, 0)
	for idx := range t.NumField() {
		field := t.Field(idx)
		errList = append(errList, collectTypeErrorsForTag(field, "param", c.Param)...)
		errList = append(errList, collectTypeErrorsForTag(field, "header", c.Request().Header.Get)...)
		errList = append(errList, collectTypeErrorsForTag(field, "query", c.QueryParam)...)
		errList = append(errList, collectTypeErrorsForTag(field, "form", c.FormValue)...)
	}

	if len(errList) == 0 {
		return nil
	}

	return NewErrors(errList...)
}

func collectTypeErrorsForTag(field reflect.StructField, tag string, valueFn func(string) string) []FieldError {
	key := strings.Split(field.Tag.Get(tag), ",")[0]
	if key == "" || key == "-" {
		return nil
	}

	raw := valueFn(key)
	if raw == "" {
		return nil
	}

	fieldType := field.Type
	for fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}

	if isQueryValueCompatible(raw, fieldType) {
		return nil
	}

	return []FieldError{{
		Field:   key,
		Message: fmt.Sprintf("%s must be a %s", prettyFieldName(key), expectedTypeLabel(fieldType)),
	}}
}

func isQueryValueCompatible(raw string, t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Invalid,
		reflect.Uintptr,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice,
		reflect.Struct,
		reflect.UnsafePointer:
		return true
	case reflect.String:
		return true
	case reflect.Bool:
		_, err := strconv.ParseBool(raw)
		return err == nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err := strconv.ParseInt(raw, 10, t.Bits())
		return err == nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, err := strconv.ParseUint(raw, 10, t.Bits())
		return err == nil
	case reflect.Float32, reflect.Float64:
		_, err := strconv.ParseFloat(raw, t.Bits())
		return err == nil
	default:
		return true
	}
}

func expectedTypeLabel(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Invalid,
		reflect.Uintptr,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Pointer,
		reflect.Slice,
		reflect.Struct,
		reflect.UnsafePointer:
		return strings.ToLower(t.Kind().String())
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return typeLabelNumber
	default:
		return strings.ToLower(t.Kind().String())
	}
}

func extractUnmarshalTypeError(err error) *json.UnmarshalTypeError {
	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		return unmarshalTypeErr
	}

	return nil
}

func normalizeJSONField(field string) string {
	if field == "" {
		return "request_body"
	}
	parts := strings.Split(field, ".")
	return parts[len(parts)-1]
}

func prettyFieldName(field string) string {
	if field == "" {
		return "Field"
	}

	name := strings.ReplaceAll(field, "_", " ")
	parts := strings.Fields(name)
	for i := range parts {
		if len(parts[i]) == 0 {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
			continue
		}
		parts[i] = strings.ToLower(parts[i])
	}
	return strings.Join(parts, " ")
}
