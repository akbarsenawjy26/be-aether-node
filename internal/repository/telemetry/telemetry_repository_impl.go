package telemetry

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type telemetryRepository struct {
	influxClient influxdb2.Client
	bucket      string
	org         string
}

func NewTelemetryRepository(influxClient influxdb2.Client, org, bucket string) TelemetryRepository {
	return &telemetryRepository{
		influxClient: influxClient,
		bucket:       bucket,
		org:          org,
	}
}

func (r *telemetryRepository) WriteTelemetry(ctx context.Context, telemetry *Telemetry) error {
	point := influxdb2.NewPoint(
		"telemetry",
		map[string]string{
			"device_sn":      telemetry.DeviceSN,
			"device_type":    telemetry.DeviceType,
			"location_name":  telemetry.LocationName,
		},
		map[string]interface{}{
			"temperature":    telemetry.Temperature,
			"humidity":       telemetry.Humidity,
			"aqi":            telemetry.AirQualityIndex,
			"pm25":           telemetry.PM25,
			"pm10":           telemetry.PM10,
			"co2":            telemetry.CO2,
			"voc":            telemetry.VOC,
		},
		telemetry.Timestamp,
	)

	writeAPI := r.influxClient.WriteAPI(r.org, r.bucket)
	writeAPI.WritePoint(ctx, point)
	writeAPI.Flush()

	return nil
}

func (r *telemetryRepository) GetLatestByDeviceSN(ctx context.Context, deviceSN string) (*Telemetry, error) {
	queryAPI := r.influxClient.QueryAPI(r.org)

	query := `
		from(bucket: "` + r.bucket + `")
		|> range(start: -1h)
		|> filter(fn: (r) => r["device_sn"] == "` + deviceSN + `")
		|> last()
	`

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	// Process result - simplified for example
	// In production, you'd properly iterate over flux.Result
	telemetry := &Telemetry{
		DeviceSN: deviceSN,
		// Other fields would be populated from the result
	}

	return telemetry, nil
}

func (r *telemetryRepository) GetAllLatest(ctx context.Context) ([]*Telemetry, error) {
	queryAPI := r.influxClient.QueryAPI(r.org)

	query := `
		import "influxdata/influxdb/schema"
		from(bucket: "` + r.bucket + `")
		|> range(start: -1h)
		|> filter(fn: (r) => r["_measurement"] == "telemetry")
		|> schema.tagValues(tagKey: "device_sn")
	`

	// For simplicity, return empty array
	// In production, you'd properly query and aggregate
	return []*Telemetry{}, nil
}

func (r *telemetryRepository) QueryHistory(ctx context.Context, deviceSN string, query *TelemetryQuery) (*TelemetryListResult, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Order == "" {
		query.Order = "time"
	}
	if query.Sort == "" {
		query.Sort = "DESC"
	}

	offset := (query.Page - 1) * query.Limit

	// Parse time range
	startTime := time.Now().Add(-24 * time.Hour)
	if query.StartTime != "" {
		startTime, _ = time.Parse(time.RFC3339, query.StartTime)
	}

	endTime := time.Now()
	if query.EndTime != "" {
		endTime, _ = time.Parse(time.RFC3339, query.EndTime)
	}

	queryAPI := r.influxClient.QueryAPI(r.org)

	fluxQuery := `
		from(bucket: "` + r.bucket + `")
		|> range(start: ` + startTime.Format(time.RFC3339) + `, stop: ` + endTime.Format(time.RFC3339) + `)
		|> filter(fn: (r) => r["device_sn"] == "` + deviceSN + `")
		|> limit(n: ` + string(rune('0'+query.Limit)) + `, offset: ` + string(rune('0'+offset)) + `)
		|> sort(columns: ["` + query.Order + `"], desc: ` + (query.Sort == "DESC") + `)
	`

	// In production, you'd properly iterate over flux.Result
	// For now, return empty result
	result := &TelemetryListResult{
		Telemetry: []*Telemetry{},
		Total:     0,
		Page:      query.Page,
		Limit:     query.Limit,
		TotalPage: 0,
	}

	return result, nil
}
