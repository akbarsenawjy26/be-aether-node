package config

import (
	"os"
	"strconv"
	"time"
)

type JWTConfig struct {
	Secret               string `env:"JWT_SECRET" envDefault:""`
	AccessExpiryMinutes  int    `env:"JWT_ACCESS_EXPIRY_MINUTES" envDefault:"15"`
	RefreshExpiryDays    int    `env:"JWT_REFRESH_EXPIRY_DAYS" envDefault:"7"`
}

func (j *JWTConfig) Load() {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		j.Secret = v
	}
	if v := os.Getenv("JWT_ACCESS_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			j.AccessExpiryMinutes = n
		}
	}
	if v := os.Getenv("JWT_REFRESH_EXPIRY_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			j.RefreshExpiryDays = n
		}
	}
}

func (j *JWTConfig) AccessExpiry() time.Duration {
	return time.Duration(j.AccessExpiryMinutes) * time.Minute
}

func (j *JWTConfig) RefreshExpiry() time.Duration {
	return time.Duration(j.RefreshExpiryDays) * 24 * time.Hour
}
