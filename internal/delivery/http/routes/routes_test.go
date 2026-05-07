package routes

import (
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/labstack/echo/v5"
)

func TestRegister(t *testing.T) {
	e := echo.New()
	rc := RouteConfig{
		Echo:               e,
		Config:             config.Config{Server: config.Server{Port: 8080}},
		AuthMiddleware:     func(next echo.HandlerFunc) echo.HandlerFunc { return next },
		AuthHandler:        handler.NewAuthHandler(nil),
		ProfileHandler:     handler.NewProfileHandler(nil),
		UserHandler:        handler.NewUserHandler(nil),
		HealthCheckHandler: handler.NewHealthCheckHandler(nil, nil),
	}

	Register(rc)

	routes := e.Router().Routes()
	if len(routes) == 0 {
		t.Fatal("expected routes to be registered")
	}

	mustHave := map[string]bool{
		"GET /health":                                      false,
		"POST /api/v1/auth/register":                       false,
		"POST /api/v1/auth/login":                          false,
		"POST /api/v1/auth/refresh-token":                  false,
		"POST /api/v1/auth/forgot-password/send-otp":       false,
		"POST /api/v1/auth/forgot-password/verify-otp":     false,
		"POST /api/v1/auth/forgot-password/reset-password": false,
		"GET /api/v1/profile":                              false,
		"PUT /api/v1/profile":                              false,
		"PUT /api/v1/profile/password":                     false,
		"GET /api/v1/users":                                false,
	}

	for _, r := range routes {
		key := r.Method + " " + r.Path
		if _, ok := mustHave[key]; ok {
			mustHave[key] = true
		}
	}

	for k, ok := range mustHave {
		if !ok {
			t.Fatalf("missing route: %s", k)
		}
	}
}
