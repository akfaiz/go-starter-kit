package handler_test

import (
	"net/http/httptest"
	"strings"

	appvalidator "github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/labstack/echo/v5"
)

func newJSONContext(method, path, payload string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	v := appvalidator.New()
	e.Validator = v
	e.Binder = appvalidator.NewBinder(e.Binder, v)
	req := httptest.NewRequest(method, path, strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}
