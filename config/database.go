package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type DatabaseConfig struct {
	Host            string `env:"DATABASE_HOST" envDefault:"localhost"`
	Port            int    `env:"DATABASE_PORT" envDefault:"5432"`
	User            string `env:"DATABASE_USER" envDefault:"postgres"`
	Password        string `env:"DATABASE_PASSWORD" envDefault:"postgres"`
	Name            string `env:"DATABASE_NAME" envDefault:"aether_node"`
	MaxConn         int    `env:"DATABASE_MAX_CONN" envDefault:"25"`
	MinConn         int    `env:"DATABASE_MIN_CONN" envDefault:"5"`
	MaxConnLifetime int    `env:"DATABASE_MAX_CONN_LIFETIME" envDefault:"300"` // seconds
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

func (d *DatabaseConfig) MaxConnLifetimeDuration() time.Duration {
	return time.Duration(d.MaxConnLifetime) * time.Second
}

func (d *DatabaseConfig) Load() {
	if v := os.Getenv("DATABASE_HOST"); v != "" {
		d.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			d.Port = port
		}
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		d.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		d.Password = v
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		d.Name = v
	}
	if v := os.Getenv("DATABASE_MAX_CONN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			d.MaxConn = n
		}
	}
	if v := os.Getenv("DATABASE_MIN_CONN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			d.MinConn = n
		}
	}
	if v := os.Getenv("DATABASE_MAX_CONN_LIFETIME"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			d.MaxConnLifetime = n
		}
	}
}
