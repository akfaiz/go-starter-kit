package validator_test

import (
	"errors"
	"testing"

	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registerRequest struct {
	Email                string `json:"email" validate:"required,email" label:"Email"`
	Password             string `json:"password" validate:"required,min=8" label:"Password"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password" label:"Confirm Password"`
}

type nestedProfileRequest struct {
	Profile nestedProfile `json:"profile"`
}

type nestedProfile struct {
	Street string `json:"street" validate:"required" label:"Street"`
}

type stringArrayRequest struct {
	Tags []string `json:"tags" validate:"required,dive,required" label:"Tags"`
}

type arrayStructRequest struct {
	Phones []phone `json:"phones" validate:"required,dive" label:"Phones"`
}

type phone struct {
	Number string `json:"number" validate:"required" label:"Number"`
}

func TestValidate_Success(t *testing.T) {
	v := validator.New()
	req := &registerRequest{
		Email:                "john.doe@example.com",
		Password:             "supersecret",
		PasswordConfirmation: "supersecret",
	}

	err := v.Validate(req)
	require.NoError(t, err)
}

func TestValidate_ReturnsValidationErrorWithJSONFieldAndTranslatedMessage(t *testing.T) {
	v := validator.New()
	req := &registerRequest{
		Email:                "string",
		Password:             "supersecret",
		PasswordConfirmation: "supersecret",
	}

	err := v.Validate(req)
	require.Error(t, err)

	var vErr *validator.ValidationError
	require.True(t, errors.As(err, &vErr), "expected *ValidationError, got %T", err)
	require.Len(t, *vErr, 1)

	first := vErr.First()
	require.NotNil(t, first)
	assert.Equal(t, "email", first.Field)
	assert.Equal(t, "Email must be a valid email address", first.Message)
}

func TestValidate_UsesLabelInMessageAndJSONFieldKey(t *testing.T) {
	v := validator.New()
	req := &registerRequest{
		Email:                "john.doe@example.com",
		Password:             "",
		PasswordConfirmation: "",
	}

	err := v.Validate(req)
	require.Error(t, err)

	var vErr *validator.ValidationError
	require.True(t, errors.As(err, &vErr), "expected *ValidationError, got %T", err)

	found := map[string]string{}
	for _, fieldErr := range *vErr {
		found[fieldErr.Field] = fieldErr.Message
	}

	assert.Equal(t, "Password is a required field", found["password"])
	assert.Equal(t, "Confirm Password is a required field", found["password_confirmation"])
}

func TestValidate_NestedStructPathUsesJSONKeys(t *testing.T) {
	v := validator.New()
	req := &nestedProfileRequest{
		Profile: nestedProfile{
			Street: "",
		},
	}

	err := v.Validate(req)
	require.Error(t, err)

	var vErr *validator.ValidationError
	require.True(t, errors.As(err, &vErr), "expected *ValidationError, got %T", err)
	require.NotNil(t, vErr.First())

	assert.Equal(t, "profile.street", vErr.First().Field)
	assert.Equal(t, "Street is a required field", vErr.First().Message)
}

func TestValidate_ArrayPathUsesJSONKeysWithIndex(t *testing.T) {
	v := validator.New()
	req := &stringArrayRequest{
		Tags: []string{""},
	}

	err := v.Validate(req)
	require.Error(t, err)

	var vErr *validator.ValidationError
	require.True(t, errors.As(err, &vErr), "expected *ValidationError, got %T", err)
	require.NotNil(t, vErr.First())

	assert.Equal(t, "tags[0]", vErr.First().Field)
	assert.Equal(t, "Tags[0] is a required field", vErr.First().Message)
}

func TestValidate_ArrayStructPathUsesJSONKeysWithIndex(t *testing.T) {
	v := validator.New()
	req := &arrayStructRequest{
		Phones: []phone{
			{Number: ""},
		},
	}

	err := v.Validate(req)
	require.Error(t, err)

	var vErr *validator.ValidationError
	require.True(t, errors.As(err, &vErr), "expected *ValidationError, got %T", err)
	require.NotNil(t, vErr.First())

	assert.Equal(t, "phones[0].number", vErr.First().Field)
	assert.Equal(t, "Number is a required field", vErr.First().Message)
}
