package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr string

	BaseURL string // e.g. http://localhost:8080

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	DBMaxOpen int
	DBMaxIdle int
	DBMaxLife time.Duration
}

func FromEnv() Config {
	// best-effort .env loading (no error if missing)
	_ = godotenv.Load()

	c := Config{
		HTTPAddr: getenv("HTTP_ADDR", ":8080"),
		BaseURL:  strings.TrimRight(getenv("BASE_URL", "http://localhost:8080"), "/"),

		DBHost: getenv("DB_HOST", "mariadb"),
		DBPort: getenv("DB_PORT", "3306"),
		DBUser: getenv("DB_USER", "shortener"),
		DBPass: getenv("DB_PASSWORD", "shortener"),
		DBName: getenv("DB_NAME", "shortener"),

		DBMaxOpen: atoi(getenv("DB_MAX_OPEN_CONNS", "50"), 50),
		DBMaxIdle: atoi(getenv("DB_MAX_IDLE_CONNS", "25"), 25),
		DBMaxLife: parseDuration(getenv("DB_CONN_MAX_LIFETIME", "5m"), 5*time.Minute),
	}
	return c
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoi(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func parseDuration(s string, def time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return def
	}
	return d
}
