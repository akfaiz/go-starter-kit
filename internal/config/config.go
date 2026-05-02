package config

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	App      App
	Auth     Auth
	Database Database
	Mail     Mail
	Server   Server
}

func Load() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	return Config{
		App:      loadAppConfig(),
		Auth:     getAuthConfig(),
		Database: loadDatabaseConfig(),
		Mail:     loadMailConfig(),
		Server:   loadServerConfig(),
	}
}
