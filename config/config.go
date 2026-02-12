package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv      string
	AppPort     string
	AppMode     string
	Prefork     bool
	DBHost      string
	DBPort      string
	DBUser      string
	DBPass      string
	DBName      string
	EmailServerAPIKey string
	EmailSMTPHost     string
	EmailSMTPUser     string
	EmailSMTPPass     string
	EmailSMTPMailFrom string
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

	emailServerAPIKey := os.Getenv("EMAIL_SERVER_API_KEY")
	emailSMTPHost := os.Getenv("EMAIL_SMTP_HOST")
	emailSMTPUser := os.Getenv("EMAIL_SMTP_USER")
	emailSMTPPass := os.Getenv("EMAIL_SMTP_PASS")
	emailSMTPMailFrom := os.Getenv("EMAIL_SMTP_MAIL_FROM")

	return &Config{
		AppEnv:            appEnv,
		AppPort:           appPort,
		AppMode:           appMode,
		Prefork:           prefork,
		DBHost:            dbHost,
		DBPort:            dbPort,
		DBUser:            dbUser,
		DBPass:            dbPass,
		DBName:            dbName,
		EmailServerAPIKey: emailServerAPIKey,
		EmailSMTPHost:     emailSMTPHost,
		EmailSMTPUser:     emailSMTPUser,
		EmailSMTPPass:     emailSMTPPass,
		EmailSMTPMailFrom: emailSMTPMailFrom,
	}, nil
}
