package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string

	ServerHost string
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	LogLevel string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return &Config{
		AppName:    getEnv("APP_NAME", "agri-api"),
		AppEnv:     getEnv("APP_ENV", "development"),
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "agri"),
		DBPassword: getEnv("DB_PASSWORD", "agri123"),
		DBName:     getEnv("DB_NAME", "agri_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
