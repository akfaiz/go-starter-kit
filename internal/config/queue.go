package config

import (
	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Queue struct {
	Concurrency    int
	Strict         bool
	MetricsEnabled bool
	MetricsPort    int
}

func loadQueueConfig() Queue {
	return Queue{
		Concurrency:    env.GetInt("QUEUE_CONCURRENCY", 10),
		Strict:         env.GetBool("QUEUE_STRICT", false),
		MetricsEnabled: env.GetBool("QUEUE_METRICS_ENABLED", true),
		MetricsPort:    env.GetInt("QUEUE_METRICS_PORT", 8081),
	}
}
