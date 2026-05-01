package telemetry

import "context"

type TelemetryService interface {
	// WriteTelemetry writes telemetry data from a device
	WriteTelemetry(ctx context.Context, telemetry *Telemetry) error

	// StreamAllDevices returns a channel of telemetry from all devices
	StreamAllDevices(ctx context.Context) (<-chan *Telemetry, <-chan error)

	// StreamDevice returns a channel of telemetry for a specific device
	StreamDevice(ctx context.Context, deviceSN string) (<-chan *Telemetry, <-chan error)

	// GetHistory retrieves historical telemetry data
	GetHistory(ctx context.Context, deviceSN string, query *TelemetryQuery) (*TelemetryListResult, error)
}
