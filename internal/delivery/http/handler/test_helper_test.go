package handler_test

import (
	"net/http"

	"github.com/akfaiz/go-starter-kit/internal/delivery/http/server"
	appvalidator "github.com/akfaiz/go-starter-kit/pkg/validator"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
)

func newExpect(e *echo.Echo) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		Client: &http.Client{
			Transport: httpexpect.NewBinder(e),
			Jar:       httpexpect.NewCookieJar(),
		},
		Reporter: httpexpect.NewRequireReporter(GinkgoT()),
	})
}

func setupEcho() *echo.Echo {
	e := echo.New()
	v := appvalidator.New()
	e.Validator = v
	e.Binder = appvalidator.NewBinder(e.Binder, v)
	e.HTTPErrorHandler = server.CustomHTTPErrorHandler
	return e
}
