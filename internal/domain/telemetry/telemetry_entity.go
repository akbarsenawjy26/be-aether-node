package telemetry

import "time"

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
