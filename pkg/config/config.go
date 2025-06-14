package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds
type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	ServerPort      string
	WebhookURL      string
	WebhookAuthKey  string
	DefaultPageSize int
	MaxPageSize     int
}

// Load loads configuration from env
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	return &Config{
		DBHost:          getEnv("DB_HOST", "postgres"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "postgres"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		WebhookURL:      getEnv("WEBHOOK_URL", ""),
		WebhookAuthKey:  getEnv("WEBHOOK_AUTH_KEY", ""),
		DefaultPageSize: getEnvAsInt("DEFAULT_PAGE_SIZE", 10),
		MaxPageSize:     getEnvAsInt("MAX_PAGE_SIZE", 100),
	}, nil
}

// getEnv retrieves an environment variable value or return default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or return default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
