package config

import (
	"os"
	"strings"
)

type NATSConfig struct {
	URL string
}

func (c *NATSConfig) Load() {
	// Ambil dari environment variable
	url := os.Getenv("NATS_URL")
	
	// Bersihkan jika ada whitespace atau tanda kutip
	url = strings.Trim(url, "\" ")
	url = strings.TrimSpace(url)

	if url == "" {
		url = "nats://localhost:4222"
	}
	
	c.URL = url
}
