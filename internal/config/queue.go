package config

import (
	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Queue struct {
	Concurrency int
	Strict      bool
}

func loadQueueConfig() Queue {
	return Queue{
		Concurrency: env.GetInt("QUEUE_CONCURRENCY", 10),
		Strict:      env.GetBool("QUEUE_STRICT", false),
	}
}
