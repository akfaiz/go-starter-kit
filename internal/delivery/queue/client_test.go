package queue_test

import (
	"context"
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/config"
	"github.com/akfaiz/go-starter-kit/internal/delivery/queue"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := queue.NewClient(config.Redis{Addr: mr.Addr()})
	require.NotNil(t, client)
	require.NoError(t, client.Close())
}

func TestClient_EnqueueContext(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := queue.NewClient(config.Redis{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	task := queue.NewTask(context.Background(), queue.TypeMailSend, []byte(`{"to":["test@example.com"]}`))

	info, err := client.EnqueueContext(context.Background(), task, asynq.Queue("mail"))
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, queue.TypeMailSend, info.Type)
	assert.Equal(t, "mail", info.Queue)
	assert.Equal(t, task.Payload(), info.Payload)

	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = inspector.Close()
	})

	pending, err := inspector.ListPendingTasks("mail")
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, info.ID, pending[0].ID)
	assert.Equal(t, queue.TypeMailSend, pending[0].Type)
	assert.Equal(t, task.Payload(), pending[0].Payload)
}
