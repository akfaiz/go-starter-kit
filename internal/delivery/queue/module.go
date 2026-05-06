package queue

import (
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/handler"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/middleware"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/server"
	"github.com/hibiken/asynq"
	"go.uber.org/fx"
)

var ClientModule = fx.Module("queue_client",
	fx.Provide(
		NewClient,
	),
	fx.Invoke(server.RegisterQueueMetrics),
)

var WorkerModule = fx.Module("queue_worker",
	fx.Provide(
		asynq.NewServeMux,
		handler.NewMailTaskHandler,
		server.NewServer,
	),
	fx.Invoke(Register),
	fx.Invoke(server.RegisterServer),
)

type WorkerConfig struct {
	fx.In

	Mux             *asynq.ServeMux
	MailTaskHandler *handler.MailTaskHandler
}

func Register(cfg WorkerConfig) {
	cfg.Mux.Use(middleware.Otel)
	cfg.Mux.Use(middleware.Logger)

	cfg.Mux.HandleFunc(TypeMailSend, cfg.MailTaskHandler.HandleSendMail)
}
