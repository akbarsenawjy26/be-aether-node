package telemetry

import (
	"context"
	"time"
)

// ============================================================
// Existing Types (kept for write operations)
// ============================================================

type Telemetry struct {
	DeviceSN         string    `json:"device_sn"`
	DeviceType       string    `json:"device_type,omitempty"`
	LocationName     string    `json:"location_name,omitempty"`
	Temperature      float64   `json:"temperature"`
	Humidity         float64   `json:"humidity"`
	AirQualityIndex  int       `json:"aqi"`
	PM25             float64   `json:"pm25"`
	PM10             float64   `json:"pm10"`
	CO2              int       `json:"co2"`
	VOC              float64   `json:"voc"`
	Timestamp        time.Time `json:"timestamp"`
}

type TelemetryQuery struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Limit     int    `json:"limit"`
	Page      int    `json:"page"`
	Order     string `json:"order"`
	Sort      string `json:"sort"`
}

type TelemetryListResult struct {
	Telemetry []*Telemetry
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

// ============================================================
// New Types (for SSE streaming + history from DEVICE_STREAM_GUIDE)
// ============================================================

// HealthData — data kesehatan device dari InfluxDB measurement "health"
type HealthData struct {
	DeviceSN    string    `json:"device_sn"`
	GatewaySN   string    `json:"gateway_sn"`
	Project     string    `json:"project"`
	Type        string    `json:"type"`
	Model       string    `json:"model"`
	Status      string    `json:"status"`       // "online" | "offline"
	Uptime      float64   `json:"uptime"`
	Temp        float64   `json:"temp"`
	Hum         float64   `json:"hum"`
	RSSI        float64   `json:"rssi"`
	ResetReason string    `json:"reset_reason"`
	LastSeen    time.Time `json:"last_seen"`
}

// TelemetryData — fields dinamis per device type (map[string]interface{})
type TelemetryData map[string]interface{}

// TelemetryRecord — satu titik waktu untuk history
type TelemetryRecord struct {
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// DeviceEntry — satu device lengkap untuk SSE payload
type DeviceEntry struct {
	Health    HealthData    `json:"health"`
	Telemetry TelemetryData `json:"telemetry"`
}

// DevicePayload — SSE dashboard payload grouped by device type
// Contoh: {"basic-model": [...], "aqi-model": [...]}
type DevicePayload map[string][]DeviceEntry

// QueryTimeRange — time range untuk history query
type QueryTimeRange struct {
	Start time.Time
	Stop  time.Time
}

// DeviceFilter — filter untuk query latest (SSE)
type DeviceFilter struct {
	Project  string
	DeviceSN string // kosong = semua device
}

// HistoryFilter — filter untuk query history
type HistoryFilter struct {
	Project  string
	DeviceSN string
	TimeRange QueryTimeRange
	Window   string // "1m", "5m", "1h" — kosong = raw data
}

// ============================================================
// Repository Interface (updated)
// ============================================================

type TelemetryRepository interface {
	// Existing write operation
	WriteTelemetry(ctx context.Context, telemetry *Telemetry) error

	// Existing read operations (keep for backward compatibility)
	GetLatestByDeviceSN(ctx context.Context, deviceSN string) (*Telemetry, error)
	GetAllLatest(ctx context.Context) ([]*Telemetry, error)
	QueryHistory(ctx context.Context, deviceSN string, query *TelemetryQuery) (*TelemetryListResult, error)

	// NEW: Health + Telemetry split queries (for SSE streaming)
	GetLatestHealth(ctx context.Context, filter DeviceFilter) ([]HealthData, error)
	GetLatestTelemetry(ctx context.Context, filter DeviceFilter) (map[string]TelemetryData, error)
	GetTelemetryHistory(ctx context.Context, filter HistoryFilter) ([]TelemetryRecord, error)

	// SSE Bypass: same queries but skip circuit breaker for long-lived connections
	GetLatestHealthSSE(ctx context.Context, filter DeviceFilter) ([]HealthData, error)
	GetLatestTelemetrySSE(ctx context.Context, filter DeviceFilter) (map[string]TelemetryData, error)
}

// ============================================================
// Service Interface (updated)
// ============================================================

type TelemetryService interface {
	// Existing write operation
	WriteTelemetry(ctx context.Context, telemetry *Telemetry) error

	// Existing streaming operations (keep for backward compatibility)
	StreamAllDevices(ctx context.Context) (<-chan *Telemetry, <-chan error)
	StreamDevice(ctx context.Context, deviceSN string) (<-chan *Telemetry, <-chan error)
	GetHistory(ctx context.Context, deviceSN string, query *TelemetryQuery) (*TelemetryListResult, error)

	// NEW: SSE streaming with health + telemetry merged
	StreamAllDevicesWithHealth(ctx context.Context, project string) (DevicePayload, error)
	StreamDeviceWithHealth(ctx context.Context, project, deviceSN string) (*DeviceEntry, error)
	GetTelemetryHistory(ctx context.Context, filter HistoryFilter) ([]TelemetryRecord, error)
}
