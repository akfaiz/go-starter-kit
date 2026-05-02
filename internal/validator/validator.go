package validator

import "github.com/gookit/validate"

type Validate struct{}

func New() *Validate {
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = false
	})

	return &Validate{}
}

func (v *Validate) Validate(i interface{}) error {
	val := validate.Struct(i)
	val.Validate()
	if val.Errors.Empty() {
		return nil
	}
	fieldErrors := make([]FieldError, 0, len(val.Errors))
	for field, ms := range val.Errors {
		for _, m := range ms {
			fieldErrors = append(fieldErrors, FieldError{
				Field:   field,
				Message: m,
			})
			break // only take the first error message per field
		}
	}
	return NewErrors(fieldErrors...)
}
