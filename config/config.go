package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort                string
	AuthServiceBaseURL     string
	AuthServiceGRPCAddr    string
	DefaultLang            string
	DBHost                 string
	DBPort                 string
	DBUser                 string
	DBPassword             string
	DBName                 string
	AdminAPIToken          string
	CORSAllowOrigins       []string
	LogRetentionDays       int
	TrustProxyHeaders      bool
	AllowLoopbackUpstreams bool
	JWTSecret              string
	JWTPublicKeyPEM        string
	JWTIssuer              string
	JWTAudience            string
	APIKeys                []string
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
			AppPort:                getEnv("APP_PORT", "8080"),
			AuthServiceBaseURL:     os.Getenv("AUTH_SERVICE_BASE_URL"),
			AuthServiceGRPCAddr:    os.Getenv("AUTH_SERVICE_GRPC_ADDR"),
			DefaultLang:            getEnv("DEFAULT_LANG", "id"),
			DBHost:                 getEnv("DB_HOST", "localhost"),
			DBPort:                 getEnv("DB_PORT", "5432"),
			DBUser:                 os.Getenv("DB_USER"),
			DBPassword:             os.Getenv("DB_PASSWORD"),
			DBName:                 os.Getenv("DB_NAME"),
			AdminAPIToken:          os.Getenv("ADMIN_API_TOKEN"),
			LogRetentionDays:       getEnvInt("LOG_RETENTION_DAYS", 7),
			TrustProxyHeaders:      getEnv("TRUST_PROXY_HEADERS", "true") != "false",
			AllowLoopbackUpstreams: getEnv("ALLOW_LOOPBACK_UPSTREAMS", "false") == "true",
			JWTSecret:              os.Getenv("JWT_SECRET"),
			JWTPublicKeyPEM:        os.Getenv("JWT_PUBLIC_KEY_PEM"),
			JWTIssuer:              os.Getenv("JWT_ISSUER"),
			JWTAudience:            os.Getenv("JWT_AUDIENCE"),
			APIKeys:                splitCSVOrNil(os.Getenv("API_KEYS")),
			CORSAllowOrigins:       splitCSV(getEnv("CORS_ALLOW_ORIGINS", "*")),
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

func splitCSV(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func getEnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

func splitCSVOrNil(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Reset discards the cached config so the next Load re-reads the environment.
// Intended for tests.
func Reset() {
	instance = nil
	once = sync.Once{}
}
