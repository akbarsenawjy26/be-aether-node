package telemetry

import (
	"context"
)

type TelemetryRepository interface {
	// WriteTelemetry writes a telemetry data point to InfluxDB
	WriteTelemetry(ctx context.Context, telemetry *Telemetry) error

	// GetLatestByDeviceSN retrieves the latest telemetry for a device
	GetLatestByDeviceSN(ctx context.Context, deviceSN string) (*Telemetry, error)

	// GetAllLatest retrieves the latest telemetry for all active devices
	GetAllLatest(ctx context.Context) ([]*Telemetry, error)

	// QueryHistory retrieves historical telemetry data
	QueryHistory(ctx context.Context, deviceSN string, query *TelemetryQuery) (*TelemetryListResult, error)
}
