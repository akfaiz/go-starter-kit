package config

import "github.com/akfaiz/go-starter-kit/pkg/env"

type Redis struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
}

func loadRedisConfig() Redis {
	return Redis{
		Addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
		Password: env.GetString("REDIS_PASSWORD"),
		DB:       env.GetInt("REDIS_DB", 0),
		Prefix:   env.GetString("REDIS_PREFIX", "gsk"),
	}
}
