package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string
	AppMode string
	Prefork bool
	DBHost  string
	DBPort  string
	DBUser  string
	DBPass  string
	DBName  string
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

	appMode := os.Getenv("APP_MODE")
	if appMode == "" {
		appMode = "web"
	}

	prefork := appEnv == "production"

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "edvance"
	}

	return &Config{
		AppEnv:  appEnv,
		AppPort: appPort,
		AppMode: appMode,
		Prefork: prefork,
		DBHost:  dbHost,
		DBPort:  dbPort,
		DBUser:  dbUser,
		DBPass:  dbPass,
		DBName:  dbName,
	}, nil
}
