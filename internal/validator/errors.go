package validator

import (
	"fmt"
	"strings"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError []FieldError

func NewError(field, message string) *ValidationError {
	return &ValidationError{{Field: field, Message: message}}
}

func NewErrors(fieldErrors ...FieldError) *ValidationError {
	validationErr := ValidationError(fieldErrors)
	return &validationErr
}

func (ve ValidationError) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return fmt.Sprintf("validation error: %s", strings.Join(messages, ", "))
}

func (ve ValidationError) Errors() []FieldError {
	return ve
}

func (ve ValidationError) First() *FieldError {
	if len(ve) == 0 {
		return nil
	}
	return &ve[0]
}

func (ve *ValidationError) Fields() []string {
	fields := make([]string, len(*ve))
	for i, err := range *ve {
		fields[i] = err.Field
	}
	return fields
}

func (ve *ValidationError) Messages() []string {
	messages := make([]string, len(*ve))
	for i, err := range *ve {
		messages[i] = err.Message
	}
	return messages
}

func (ve *ValidationError) Add(field, message string) *ValidationError {
	*ve = append(*ve, FieldError{Field: field, Message: message})
	return ve
}

func (ve *ValidationError) Addf(field, format string, args ...any) *ValidationError {
	message := fmt.Sprintf(format, args...)
	*ve = append(*ve, FieldError{Field: field, Message: message})
	return ve
}
