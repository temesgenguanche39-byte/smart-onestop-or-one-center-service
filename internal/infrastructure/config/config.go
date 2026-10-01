package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseDriver    string // "postgres" or "sqlite"
	DatabaseURL       string
	RedisURL          string
	JWTSecret         string
	JWTExpiration     time.Duration
	WorkerIntervalSec int
	Environment       string
	CORSAllowOrigins  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "sqlite")
	dbURL := getEnv("DATABASE_URL", "smart_onestop.db")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	jwtSecret := getEnv("JWT_SECRET", "smart-onestop-municipal-hearing-secure-token-secret-2026")
	env := getEnv("APP_ENV", "development")
	corsOrigins := getEnv("CORS_ALLOW_ORIGINS", "*")

	workerInterval, err := strconv.Atoi(getEnv("WORKER_INTERVAL_SEC", "60"))
	if err != nil {
		workerInterval = 60
	}

	return &Config{
		Port:              port,
		DatabaseDriver:    dbDriver,
		DatabaseURL:       dbURL,
		RedisURL:          redisURL,
		JWTSecret:         jwtSecret,
		JWTExpiration:     24 * time.Hour,
		WorkerIntervalSec: workerInterval,
		Environment:       env,
		CORSAllowOrigins:  corsOrigins,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
