package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string

	ServerHost   string
	ServerPort   string
	ClientOrigin string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	LogLevel string

	JWTSecret       string
	AccessTokenTTL  string
	RefreshTokenTTL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	PublicURL         string
	OpenWeatherAPIKey string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using system env")
	}

	return &Config{
		AppName:           getEnv("APP_NAME", "agri-api"),
		AppEnv:            getEnv("APP_ENV", "development"),
		ServerHost:        getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		ClientOrigin:      getEnv("CLIENT_ORIGIN", "http://localhost:3000"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "agri"),
		DBPassword:        getEnv("DB_PASSWORD", "agri123"),
		DBName:            getEnv("DB_NAME", "agri_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		JWTSecret:         getEnv("JWT_SECRET", "super-secret-key"),
		AccessTokenTTL:    getEnv("ACCESS_TOKEN_TTL", "15m"),
		RefreshTokenTTL:   getEnv("REFRESH_TOKEN_TTL", "7d"),
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getEnvInt("REDIS_DB", 0),
		SMTPHost:          getEnv("SMTP_HOST", "localhost"),
		SMTPPort:          getEnvInt("SMTP_PORT", 1025),
		SMTPUser:          getEnv("SMTP_USER", ""),
		SMTPPassword:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:          getEnv("SMTP_FROM", "noreply@agri-manager.local"),
		PublicURL:         getEnv("PUBLIC_URL", "http://localhost:8080"),
		OpenWeatherAPIKey: getEnv("OPENWEATHER_API_KEY", ""),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}
