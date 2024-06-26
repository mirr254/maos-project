package config

import (
	"os"
)

type Config struct {
	DB_HOST     string
	DB_USER     string
	DB_PORT     string
	DB_PASSWORD string
	DB_NAME     string
	DB_SSL_MODE string

	SECRET_KEY     string
	BASE_URL       string
	SMTP_HOST      string
	SMTP_PORT      string
	FROM_EMAIL     string
	EMAIL_PASSWORD string 
	ENV            string
    PULUMI_ACCESS_TOKEN string
}

func LoadConfig() *Config {

    return &Config {
        DB_HOST:     getEnv("DB_HOST", "localhost"),
		DB_USER:     getEnv("DB_USER", "postgres"),
		DB_PORT:     getEnv("DB_PORT", "5432"),
		DB_PASSWORD: getEnv("DB_PASSWORD", "postgres"),
		DB_NAME:     getEnv("DB_NAME", "maosproject"),
		DB_SSL_MODE: getEnv("DB_SSL_MODE", "disable"),

        SECRET_KEY:     getEnv("SECRET_KEY", "secret"),
        BASE_URL:       getEnv("BASE_URL", "http://localhost:8080"),
        SMTP_HOST:      getEnv("SMTP_HOST", "smtp.gmail.com"),
        SMTP_PORT:      getEnv("SMTP_PORT", "587"),
        FROM_EMAIL:     getEnv("FROM_EMAIL", ""),
        EMAIL_PASSWORD: getEnv("EMAIL_PASSWORD", ""),
        ENV:            getEnv("ENV", "development"),

        PULUMI_ACCESS_TOKEN: getEnv("PULUMI_ACCESS_TOKEN", ""),
    }
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
