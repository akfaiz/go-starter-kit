package http

import (
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/handler"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/middleware"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/routes"
	"github.com/akfaiz/go-starter-kit/internal/delivery/http/server"
	"go.uber.org/fx"
)

var Module = fx.Module("http",
	fx.Provide(
		server.New,
		middleware.New,
	),
	fx.Invoke(routes.Register),
	handler.Module,
)
