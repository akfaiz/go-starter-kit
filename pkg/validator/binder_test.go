package validator_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akfaiz/go-starter-kit/pkg/problem"
	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockBinder struct {
	mock.Mock
}

func (m *mockBinder) Bind(c *echo.Context, i any) error {
	args := m.Called(c, i)
	return args.Error(0)
}

type testRequest struct {
	Name string `json:"name" validate:"required"`
}

type binderQueryRequest struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

type binderFormRequest struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type binderParamRequest struct {
	UserID int `param:"user_id"`
}

type binderHeaderRequest struct {
	RetryAfter int `header:"X-Retry-After"`
}

func TestBinder_Bind(t *testing.T) {
	e := echo.New()
	v := validator.New()

	t.Run("binding error", func(t *testing.T) {
		mb := new(mockBinder)
		b := validator.NewBinder(mb, v)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		c := e.NewContext(req, nil)
		tr := &testRequest{}

		mb.On("Bind", c, tr).Return(errors.New("bind error"))

		err := b.Bind(c, tr)
		assert.Error(t, err)
		assert.Equal(t, "Bad Request: bind error", err.Error())
	})

	t.Run("validation error", func(t *testing.T) {
		mb := new(mockBinder)
		b := validator.NewBinder(mb, v)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		c := e.NewContext(req, nil)
		tr := &testRequest{Name: ""} // Invalid

		mb.On("Bind", c, tr).Return(nil)

		err := b.Bind(c, tr)
		assert.Error(t, err)
		var vErr *validator.ValidationError
		assert.ErrorAs(t, err, &vErr)
	})

	t.Run("success", func(t *testing.T) {
		mb := new(mockBinder)
		b := validator.NewBinder(mb, v)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		c := e.NewContext(req, nil)
		tr := &testRequest{Name: "John"}

		mb.On("Bind", c, tr).Return(nil)

		err := b.Bind(c, tr)
		assert.NoError(t, err)
	})

	t.Run("json type mismatch returns structured bad request", func(t *testing.T) {
		defaultBinder := &echo.DefaultBinder{}
		b := validator.NewBinder(defaultBinder, v)
		body := `{"name":123}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, nil)
		tr := &testRequest{}

		err := b.Bind(c, tr)
		require.Error(t, err)

		var pErr *problem.Error
		require.ErrorAs(t, err, &pErr)
		assert.Equal(t, http.StatusBadRequest, pErr.Status)
		assert.Equal(t, "One or more fields have invalid formats.", pErr.Detail)

		var vErr *validator.ValidationError
		require.IsType(t, &validator.ValidationError{}, pErr.Errors)
		vErr = pErr.Errors.(*validator.ValidationError)
		require.Len(t, *vErr, 1)
		assert.Equal(t, "name", (*vErr)[0].Field)
		assert.Equal(t, "Name must be a string", (*vErr)[0].Message)
	})

	t.Run("query type mismatch returns structured bad request with all invalid query fields", func(t *testing.T) {
		defaultBinder := &echo.DefaultBinder{}
		b := validator.NewBinder(defaultBinder, v)
		req := httptest.NewRequest(http.MethodGet, "/?page=invalid&limit=invalid", nil)
		c := e.NewContext(req, nil)
		qr := &binderQueryRequest{}

		err := b.Bind(c, qr)
		require.Error(t, err)

		var pErr *problem.Error
		require.ErrorAs(t, err, &pErr)
		assert.Equal(t, http.StatusBadRequest, pErr.Status)
		assert.Equal(t, "One or more fields have invalid formats.", pErr.Detail)

		vErr, ok := pErr.Errors.(*validator.ValidationError)
		require.True(t, ok)
		require.Len(t, *vErr, 2)
		assert.Equal(t, "page", (*vErr)[0].Field)
		assert.Equal(t, "Page must be a number", (*vErr)[0].Message)
		assert.Equal(t, "limit", (*vErr)[1].Field)
		assert.Equal(t, "Limit must be a number", (*vErr)[1].Message)
	})

	t.Run("form type mismatch returns structured bad request with all invalid form fields", func(t *testing.T) {
		defaultBinder := &echo.DefaultBinder{}
		b := validator.NewBinder(defaultBinder, v)
		body := "page=invalid&limit=invalid"
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		c := e.NewContext(req, nil)
		fr := &binderFormRequest{}

		err := b.Bind(c, fr)
		require.Error(t, err)

		var pErr *problem.Error
		require.ErrorAs(t, err, &pErr)
		assert.Equal(t, http.StatusBadRequest, pErr.Status)
		assert.Equal(t, "One or more fields have invalid formats.", pErr.Detail)

		vErr, ok := pErr.Errors.(*validator.ValidationError)
		require.True(t, ok)
		require.Len(t, *vErr, 2)
		assert.Equal(t, "page", (*vErr)[0].Field)
		assert.Equal(t, "Page must be a number", (*vErr)[0].Message)
		assert.Equal(t, "limit", (*vErr)[1].Field)
		assert.Equal(t, "Limit must be a number", (*vErr)[1].Message)
	})

	t.Run("param type mismatch returns structured bad request", func(t *testing.T) {
		defaultBinder := &echo.DefaultBinder{}
		b := validator.NewBinder(defaultBinder, v)
		req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
		c := e.NewContext(req, nil)
		c.SetPath("/users/:user_id")
		c.SetPathValues(echo.PathValues{{Name: "user_id", Value: "invalid"}})
		pr := &binderParamRequest{}

		err := b.Bind(c, pr)
		require.Error(t, err)

		var pErr *problem.Error
		require.ErrorAs(t, err, &pErr)
		assert.Equal(t, http.StatusBadRequest, pErr.Status)
		assert.Equal(t, "One or more fields have invalid formats.", pErr.Detail)

		vErr, ok := pErr.Errors.(*validator.ValidationError)
		require.True(t, ok)
		require.Len(t, *vErr, 1)
		assert.Equal(t, "user_id", (*vErr)[0].Field)
		assert.Equal(t, "User id must be a number", (*vErr)[0].Message)
	})

	t.Run("header type mismatch returns structured bad request", func(t *testing.T) {
		defaultBinder := &echo.DefaultBinder{}
		b := validator.NewBinder(defaultBinder, v)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Retry-After", "invalid")
		c := e.NewContext(req, nil)
		hr := &binderHeaderRequest{}

		err := b.Bind(c, hr)
		require.Error(t, err)

		var pErr *problem.Error
		require.ErrorAs(t, err, &pErr)
		assert.Equal(t, http.StatusBadRequest, pErr.Status)
		assert.Equal(t, "One or more fields have invalid formats.", pErr.Detail)

		vErr, ok := pErr.Errors.(*validator.ValidationError)
		require.True(t, ok)
		require.Len(t, *vErr, 1)
		assert.Equal(t, "X-Retry-After", (*vErr)[0].Field)
		assert.Equal(t, "X-Retry-After must be a number", (*vErr)[0].Message)
	})
}
