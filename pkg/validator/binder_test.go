package validator_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		assert.Equal(t, "bind error", err.Error())
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
}
