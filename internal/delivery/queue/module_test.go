package queue

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/handler"
	"github.com/hibiken/asynq"
)

func TestRegister(t *testing.T) {
	mux := asynq.NewServeMux()
	h := &handler.MailTaskHandler{}

	Register(WorkerConfig{Mux: mux, MailTaskHandler: h})

	if err := mux.ProcessTask(context.Background(), asynq.NewTask(TypeMailSend, []byte("{"))); err == nil {
		t.Fatal("expected handler to return error for invalid payload")
	}
}
