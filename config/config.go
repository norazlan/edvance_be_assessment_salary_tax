package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string
	Prefork bool
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "3000"
	}

	prefork := appEnv == "production"

	return &Config{
		AppEnv:  appEnv,
		AppPort: appPort,
		Prefork: prefork,
	}, nil
}
