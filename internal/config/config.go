package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	DatabaseURL    string
	FrontendOrigin string
	GoogleClientID string
	SessionTTL     time.Duration
}

func Load() Config {
	return Config{
		Port:           getEnvOrFiles("PORT", "8080", ".env"),
		DatabaseURL:    getEnvOrFiles("DATABASE_URL", "", ".env"),
		FrontendOrigin: getEnvOrFiles("FRONTEND_ORIGIN", "http://localhost:3000", ".env"),
		GoogleClientID: loadGoogleClientID(),
		SessionTTL:     getDurationHours("SESSION_TTL_HOURS", 24, ".env"),
	}
}

func loadGoogleClientID() string {
	return getEnvOrFiles("GOOGLE_CLIENT_ID", "", ".env", "../zorvyn/.env")
}

func getEnvOrFiles(key, fallback string, envPaths ...string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	for _, envPath := range envPaths {
		value, ok := lookupEnvFile(envPath, key)
		if ok {
			return value
		}
	}

	return fallback
}

func getDurationHours(key string, fallback int, envPaths ...string) time.Duration {
	value := getEnvOrFiles(key, "", envPaths...)
	if value != "" {
		hours, err := strconv.Atoi(value)
		if err == nil && hours > 0 {
			return time.Duration(hours) * time.Hour
		}
	}

	return time.Duration(fallback) * time.Hour
}

func lookupEnvFile(envPath, key string) (string, bool) {
	file, err := os.Open(filepath.Clean(envPath))
	if err != nil {
		return "", false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parsedKey, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		if strings.TrimSpace(parsedKey) == key {
			return strings.Trim(strings.TrimSpace(value), "\""), true
		}
	}

	return "", false
}
