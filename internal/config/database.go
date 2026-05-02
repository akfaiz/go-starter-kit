package config

import (
	"fmt"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d Database) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.SSLMode,
	)
}

func loadDatabaseConfig() Database {
	return Database{
		Host:     env.GetString("DB_HOST", "localhost"),
		Port:     env.GetInt("DB_PORT", 5432),
		User:     env.GetString("DB_USER"),
		Password: env.GetString("DB_PASSWORD"),
		Name:     env.GetString("DB_NAME"),
		SSLMode:  env.GetString("DB_SSLMODE", "disable"),
	}
}
