package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type ServerConfig struct {
	Host              string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port              int    `env:"SERVER_PORT" envDefault:"8080"`
	ShutdownTimeout   int    `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"30"` // seconds
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:""`
}

func (s *ServerConfig) Load() {
	if v := os.Getenv("SERVER_HOST"); v != "" {
		s.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			s.Port = port
		}
	}
	if v := os.Getenv("SERVER_SHUTDOWN_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			s.ShutdownTimeout = n
		}
	}
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		s.CORSAllowedOrigins = v
	}
}

func (s *ServerConfig) Address() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (s *ServerConfig) ShutdownTimeoutDuration() time.Duration {
	return time.Duration(s.ShutdownTimeout) * time.Second
}

func (s *ServerConfig) CORSOrigins() []string {
	if s.CORSAllowedOrigins == "" {
		return []string{}
	}
	origins := strings.Split(s.CORSAllowedOrigins, ",")
	result := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			result = append(result, o)
		}
	}
	return result
}
