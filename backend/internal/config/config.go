package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	FrontendURLs []string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		db = "postgres://taskforge:taskforge@localhost:5436/taskforge?sslmode=disable"
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	var front []string
	if f := os.Getenv("FRONTEND_URL"); f != "" {
		front = strings.Split(f, ",")
	}
	return Config{Port: port, DatabaseURL: db, JWTSecret: secret, FrontendURLs: front}
}

func EnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
