package helpers

import (
	"fmt"
	"os"
)

// GetEnvRequired retrieves an environment variable or returns an error if not set
func GetEnvRequired(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s not set", key)
	}
	return value, nil
}

// GetEnvOrDefault retrieves an environment variable or returns the default value if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
