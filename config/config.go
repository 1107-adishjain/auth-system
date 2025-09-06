package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	RedisAddr      string
	ServerPort     string
	JWTSecret      string
	AccessTokenTTL int
	RefreshTokenTTL int
}

func LoadConfig() *Config {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found")
		}
	}

	accessTokenTTL, _ := strconv.Atoi(getEnv("ACCESS_TOKEN_TTL", "15")) // in minutes
	refreshTokenTTL, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_TTL", "10080")) // in minutes (7 days)

	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "user"),
		DBPassword:     getEnv("DB_PASSWORD", "password"),
		DBName:         getEnv("DB_NAME", "auth_db"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		JWTSecret:      getEnv("JWT_SECRET", "a-very-secret-key"),
		AccessTokenTTL: accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
