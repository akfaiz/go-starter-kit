package server

import (
	"log/slog"
	"net/http"

	appmiddleware "github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware"
	"github.com/akfaiz/go-starter-kit/internal/validator"
	"github.com/labstack/echo/v5"
	echomiddleware "github.com/labstack/echo/v5/middleware"
)

func New() *echo.Echo {
	e := echo.New()
	e.Validator = validator.New()
	e.HTTPErrorHandler = customHTTPErrorHandler
	e.Pre(echomiddleware.RemoveTrailingSlash())
	e.Use(appmiddleware.Logger(slog.Default()))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
	}))
	e.Use(appmiddleware.I18n())
	return e
}
