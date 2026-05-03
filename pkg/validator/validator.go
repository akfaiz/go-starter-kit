package validator

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	govalidator "github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	idTranslations "github.com/go-playground/validator/v10/translations/id"
	"github.com/invopop/ctxi18n"
)

type Validate struct {
	validate *govalidator.Validate
	uni      *ut.UniversalTranslator
}

func New() *Validate {
	v := govalidator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if label := fld.Tag.Get("label"); label != "" {
			return label
		}
		for _, tag := range []string{"json", "query", "param", "form"} {
			if name := fld.Tag.Get(tag); name != "" {
				name = strings.Split(name, ",")[0]
				if name != "" && name != "-" {
					return name
				}
			}
		}
		return fld.Name
	})

	english := en.New()
	indonesian := id.New()
	uni := ut.New(english, english, indonesian)

	enTrans, _ := uni.GetTranslator("en")
	_ = enTranslations.RegisterDefaultTranslations(v, enTrans)

	idTrans, _ := uni.GetTranslator("id")
	_ = idTranslations.RegisterDefaultTranslations(v, idTrans)

	return &Validate{validate: v, uni: uni}
}

func (v *Validate) Validate(i any) error {
	return v.ValidateContext(context.Background(), i)
}

func (v *Validate) ValidateContext(ctx context.Context, i any) error {
	err := v.validate.StructCtx(ctx, i)
	if err == nil {
		return nil
	}

	var validationErrors govalidator.ValidationErrors
	ok := errors.As(err, &validationErrors)
	if !ok {
		return err
	}

	locale := "en"
	if loc := ctxi18n.Locale(ctx); loc != nil {
		locale = loc.Code().String()
	}
	trans, _ := v.uni.FindTranslator(locale)

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
			Message: fieldErr.Translate(trans),
		})
	}

	return NewErrors(fieldErrors...)
}

func jsonFieldPath(i any, structNamespace, fallback string) string {
	if structNamespace == "" {
		return fallback
	}

	t := reflect.TypeOf(i)
	if t == nil {
		return fallback
	}
	if t.Kind() == reflect.Pointer {
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

		jsonName, nextType := resolveJSONNameAndType(current, name, indexSuffix)
		pathParts = append(pathParts, jsonName+indexSuffix)
		current = nextType
	}

	if len(pathParts) == 0 {
		return fallback
	}

	return strings.Join(pathParts, ".")
}

func resolveJSONNameAndType(current reflect.Type, name, indexSuffix string) (string, reflect.Type) {
	current = dereferenceType(current)
	if current.Kind() != reflect.Struct {
		return name, current
	}

	field, ok := current.FieldByName(name)
	if !ok {
		return name, current
	}

	jsonName := ""
	for _, tag := range []string{"json", "query", "param", "form"} {
		if name := field.Tag.Get(tag); name != "" {
			jsonName = strings.Split(name, ",")[0]
			if jsonName != "" && jsonName != "-" {
				break
			}
			jsonName = ""
		}
	}

	if jsonName == "" {
		jsonName = name
	}

	nextType := dereferenceType(field.Type)
	return jsonName, unwrapIndexedType(nextType, indexSuffix)
}

func dereferenceType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func unwrapIndexedType(t reflect.Type, indexSuffix string) reflect.Type {
	indexCount := strings.Count(indexSuffix, "[")
	for range indexCount {
		if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
			t = dereferenceType(t.Elem())
		}
	}
	return t
}

func splitFieldAndIndex(part string) (string, string) {
	idx := strings.Index(part, "[")
	if idx == -1 {
		return part, ""
	}
	return part[:idx], part[idx:]
}
