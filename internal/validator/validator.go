package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	govalidator "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type Validate struct {
	validate *govalidator.Validate
	trans    ut.Translator
}

func New() *Validate {
	v := govalidator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if label := fld.Tag.Get("label"); label != "" {
			return label
		}
		if jsonName := fld.Tag.Get("json"); jsonName != "" {
			name := strings.Split(jsonName, ",")[0]
			if name != "" && name != "-" {
				return name
			}
		}
		return fld.Name
	})

	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	_ = enTranslations.RegisterDefaultTranslations(v, trans)

	return &Validate{validate: v, trans: trans}
}

func (v *Validate) Validate(i interface{}) error {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(govalidator.ValidationErrors)
	if !ok {
		return err
	}

	seenFields := make(map[string]struct{}, len(validationErrors))
	fieldErrors := make([]FieldError, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		field := fieldErr.StructField()
		if _, seen := seenFields[field]; seen {
			continue
		}
		seenFields[field] = struct{}{}

		fieldErrors = append(fieldErrors, FieldError{
			Field:   jsonFieldPath(i, fieldErr.StructNamespace(), fieldErr.StructField()),
			Message: fieldErr.Translate(v.trans),
		})
	}

	return NewErrors(fieldErrors...)
}

func jsonFieldPath(i interface{}, structNamespace, fallback string) string {
	if structNamespace == "" {
		return fallback
	}

	t := reflect.TypeOf(i)
	if t == nil {
		return fallback
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fallback
	}

	parts := strings.Split(structNamespace, ".")
	if len(parts) == 0 {
		return fallback
	}

	// Skip the root struct name in namespace.
	parts = parts[1:]
	if len(parts) == 0 {
		return fallback
	}

	pathParts := make([]string, 0, len(parts))
	current := t

	for _, part := range parts {
		name, indexSuffix := splitFieldAndIndex(part)
		if name == "" {
			continue
		}

		for current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		if current.Kind() != reflect.Struct {
			pathParts = append(pathParts, name+indexSuffix)
			continue
		}

		field, ok := current.FieldByName(name)
		if !ok {
			pathParts = append(pathParts, name+indexSuffix)
			continue
		}

		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			jsonName = name
		}
		pathParts = append(pathParts, jsonName+indexSuffix)

		current = field.Type
		for current.Kind() == reflect.Ptr {
			current = current.Elem()
		}
		indexCount := strings.Count(indexSuffix, "[")
		for i := 0; i < indexCount; i++ {
			if current.Kind() == reflect.Slice || current.Kind() == reflect.Array {
				current = current.Elem()
				for current.Kind() == reflect.Ptr {
					current = current.Elem()
				}
			}
		}
	}

	if len(pathParts) == 0 {
		return fallback
	}

	return strings.Join(pathParts, ".")
}

func splitFieldAndIndex(part string) (string, string) {
	idx := strings.Index(part, "[")
	if idx == -1 {
		return part, ""
	}
	return part[:idx], part[idx:]
}
