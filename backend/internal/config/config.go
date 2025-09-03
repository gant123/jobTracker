package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DatabaseURL    string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	Port           string
	Env            string
	JWTSecret      string
	JWTExpiry      string
	AllowedOrigins string
	EncryptionKey  string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "jobtracker"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		Port:           getEnv("PORT", "8080"),
		Env:            getEnv("ENV", "development"),
		JWTSecret:      getEnv("JWT_SECRET", "change-this-secret-key"),
		JWTExpiry:      getEnv("JWT_EXPIRY", "24h"),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		EncryptionKey:  getEnv("ENCRYPTION_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
