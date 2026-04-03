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
	FrontendOrigin string
	GoogleClientID string
	SessionTTL     time.Duration
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),
		GoogleClientID: loadGoogleClientID(),
		SessionTTL:     getDurationHours("SESSION_TTL_HOURS", 24),
	}
}

func loadGoogleClientID() string {
	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		return clientID
	}

	envPath := filepath.Clean("../zorvyn/.env")
	file, err := os.Open(envPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		if strings.TrimSpace(key) == "GOOGLE_CLIENT_ID" {
			return strings.Trim(strings.TrimSpace(value), "\"")
		}
	}

	return ""
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getDurationHours(key string, fallback int) time.Duration {
	if value := os.Getenv(key); value != "" {
		hours, err := strconv.Atoi(value)
		if err == nil && hours > 0 {
			return time.Duration(hours) * time.Hour
		}
	}

	return time.Duration(fallback) * time.Hour
}
