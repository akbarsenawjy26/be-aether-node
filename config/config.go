package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// Config holds the entire application configuration.
// All config MUST be loaded via Load() before use.
type Config struct {
	Database DatabaseConfig
	InfluxDB InfluxDBConfig
	JWT      JWTConfig
	Server   ServerConfig
}

// Load reads all configuration from environment variables.
// Call this once at application startup before any other config access.
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg := &Config{}

	cfg.Database.Load()
	cfg.InfluxDB.Load()
	cfg.JWT.Load()
	cfg.Server.Load()

	// Validate required fields
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

// MustLoad loads config and panics on error.
// Use only in main() for startup config loading.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic("config load error: " + err.Error())
	}
	return cfg
}
