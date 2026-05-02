package config

import "github.com/akfaiz/go-starter-kit/pkg/env"

type Server struct {
	Port int
}

func loadServerConfig() Server {
	return Server{
		Port: env.GetInt("SERVER_PORT", 8080),
	}
}
