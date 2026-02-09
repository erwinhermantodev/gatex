package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort             string
	AuthServiceBaseURL  string
	AuthServiceGRPCAddr string
	DefaultLang         string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
}

var (
	instance *Config
	once     sync.Once
)

// Load returns the singleton config instance
func Load() *Config {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("Warning: .env file not found, using system environment variables")
		}

		instance = &Config{
			AppPort:             getEnv("APP_PORT", "8080"),
			AuthServiceBaseURL:  os.Getenv("AUTH_SERVICE_BASE_URL"),
			AuthServiceGRPCAddr: os.Getenv("AUTH_SERVICE_GRPC_ADDR"),
			DefaultLang:         getEnv("DEFAULT_LANG", "id"),
			DBHost:              getEnv("DB_HOST", "localhost"),
			DBPort:              getEnv("DB_PORT", "5432"),
			DBUser:              os.Getenv("DB_USER"),
			DBPassword:          os.Getenv("DB_PASSWORD"),
			DBName:              os.Getenv("DB_NAME"),
		}
	})
	return instance
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
