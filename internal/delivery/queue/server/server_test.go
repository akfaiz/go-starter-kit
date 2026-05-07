package server_test

import (
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue/server"
)

func TestNewServer(t *testing.T) {
	srv := server.NewServer(config.Config{
		Redis: config.Redis{Addr: "127.0.0.1:6379", DB: 0},
		Queue: config.Queue{Concurrency: 5, Strict: true},
	})

	if srv == nil {
		t.Fatal("expected non-nil server")
	}
}
