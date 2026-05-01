package config

import "os"

type InfluxDBConfig struct {
	URL    string `env:"INFLUXDB_URL" envDefault:"http://localhost:8086"`
	Token  string `env:"INFLUXDB_TOKEN" envDefault:""`
	Org    string `env:"INFLUXDB_ORG" envDefault:"aether"`
	Bucket string `env:"INFLUXDB_BUCKET" envDefault:"telemetry"`
}

func (i *InfluxDBConfig) Load() {
	if v := os.Getenv("INFLUXDB_URL"); v != "" {
		i.URL = v
	}
	if v := os.Getenv("INFLUXDB_TOKEN"); v != "" {
		i.Token = v
	}
	if v := os.Getenv("INFLUXDB_ORG"); v != "" {
		i.Org = v
	}
	if v := os.Getenv("INFLUXDB_BUCKET"); v != "" {
		i.Bucket = v
	}
}
