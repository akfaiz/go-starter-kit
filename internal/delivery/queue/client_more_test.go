package queue_test

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue"
	"github.com/alicebob/miniredis/v2"
)

func TestClient_Enqueue(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	client := queue.NewClient(config.Redis{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	task := queue.NewTask(context.Background(), queue.TypeMailSend, []byte(`{"to":["test@example.com"]}`))
	info, err := client.Enqueue(task)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if info == nil || info.Type != queue.TypeMailSend {
		t.Fatalf("unexpected task info: %#v", info)
	}
}
