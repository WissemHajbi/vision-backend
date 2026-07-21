// Package config reads runtime configuration from environment variables.
package config

import "os"

type Config struct {
	Address      string
	DatabasePath string
}

func Load() Config {
	return Config{
		Address:      envOrDefault("HTTP_ADDRESS", ":8080"),
		DatabasePath: envOrDefault("DATABASE_PATH", "./data/products.db"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
