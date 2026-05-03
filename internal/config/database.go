package config

import (
	"fmt"
	"net"
	"strconv"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Database struct {
	Host     string `validate:"required" label:"DB_HOST"`
	Port     int    `validate:"gt=0"     label:"DB_PORT"`
	User     string `validate:"required" label:"DB_USER"`
	Password string
	Name     string `validate:"required" label:"DB_NAME"`
	SSLMode  string `validate:"required" label:"DB_SSLMODE"`
}

func (d Database) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		d.User,
		d.Password,
		net.JoinHostPort(d.Host, strconv.Itoa(d.Port)),
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
