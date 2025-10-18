package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DBPath         string
	MigrationsPath string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// CORS
	CORSAllowedOrigins []string

	// Rate Limiting
	RateLimitRequests       int
	RateLimitWindowMinutes  int

	// Logging
	LogLevel string
}

// Load reads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		Port:                    getEnv("PORT", "8080"),
		Env:                     getEnv("ENV", "development"),
		DBPath:                  getEnv("DB_PATH", "./data/sprintspark.db"),
		MigrationsPath:          getEnv("MIGRATIONS_PATH", "./internal/db/migrations"),
		JWTSecret:               getEnv("JWT_SECRET", "change-this-to-a-secure-random-string-in-production"),
		JWTExpiryHours:          getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		CORSAllowedOrigins:      getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
		RateLimitRequests:       getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindowMinutes:  getEnvAsInt("RATE_LIMIT_WINDOW_MINUTES", 15),
		LogLevel:                getEnv("LOG_LEVEL", "info"),
	}

	// Validate critical configuration
	if cfg.Env == "production" && cfg.JWTSecret == "change-this-to-a-secure-random-string-in-production" {
		log.Fatal("JWT_SECRET must be set in production environment")
	}

	return cfg
}

// JWTExpiry returns the JWT expiry duration
func (c *Config) JWTExpiry() time.Duration {
	return time.Duration(c.JWTExpiryHours) * time.Hour
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

// getEnvAsSlice reads an environment variable as comma-separated values
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var result []string
	for _, v := range splitCommaSeparated(valueStr) {
		if trimmed := trimSpace(v); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}
	return result
}

// splitCommaSeparated splits a string by commas
func splitCommaSeparated(s string) []string {
	var parts []string
	current := ""
	for _, ch := range s {
		if ch == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
