package config

import (
	"fmt"

	"github.com/akfaiz/go-starter-kit/pkg/env"
)

type Database struct {
	Host         string `validate:"required" label:"DB_HOST"`
	Port         int    `validate:"gt=0"     label:"DB_PORT"`
	User         string `validate:"required" label:"DB_USER"`
	Password     string
	Name         string `validate:"required" label:"DB_NAME"`
	SSLMode      string `validate:"required" label:"DB_SSLMODE"`
	MaxOpenConns int    `validate:"gt=0"     label:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int    `validate:"gt=0"     label:"DB_MAX_IDLE_CONNS"`
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.Name,
		d.SSLMode,
	)
}

func loadDatabaseConfig() Database {
	return Database{
		Host:         env.GetString("DB_HOST", "localhost"),
		Port:         env.GetInt("DB_PORT", 5432),
		User:         env.GetString("DB_USER"),
		Password:     env.GetString("DB_PASSWORD"),
		Name:         env.GetString("DB_NAME"),
		SSLMode:      env.GetString("DB_SSLMODE", "disable"),
		MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 10),
		MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 5),
	}
}
