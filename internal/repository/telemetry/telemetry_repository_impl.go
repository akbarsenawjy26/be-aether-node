package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	domainTelemetry "aether-node/internal/domain/telemetry"
)

// InfluxDB HTTP client — direct API calls, no SDK dependency
type influxHTTPClient struct {
	url    string
	token  string
	org    string
	bucket string
	httpClient *http.Client
}

func newInfluxHTTPClient(url, token, org, bucket string) *influxHTTPClient {
	return &influxHTTPClient{
		url:    url,
		token:  token,
		org:    org,
		bucket: bucket,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type telemetryRepository struct {
	influx *influxHTTPClient
}

func NewTelemetryRepository(influxURL, token, org, bucket string) domainTelemetry.TelemetryRepository {
	return &telemetryRepository{
		influx: newInfluxHTTPClient(influxURL, token, org, bucket),
	}
}

func (r *telemetryRepository) WriteTelemetry(ctx context.Context, t *domainTelemetry.Telemetry) error {
	// Build line protocol format
	line := fmt.Sprintf(
		"telemetry,device_sn=%s,device_type=%s,location_name=%s temperature=%.2f,humidity=%.2f,aqi=%d,pm25=%.2f,pm10=%.2f,co2=%d,voc=%.2f %d",
		escapeTag(t.DeviceSN),
		escapeTag(t.DeviceType),
		escapeTag(t.LocationName),
		t.Temperature,
		t.Humidity,
		t.AirQualityIndex,
		t.PM25,
		t.PM10,
		t.CO2,
		t.VOC,
		t.Timestamp.UnixNano(),
	)

	params := url.Values{}
	params.Set("bucket", r.influx.bucket)
	params.Set("org", r.influx.org)
	params.Set("precision", "ns")

	reqURL := r.influx.url + "/api/v2/write?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(line))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+r.influx.token)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := r.influx.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("influxdb write error: status=%d body=%s", resp.StatusCode, string(body))
	}
	return nil
}

func (r *telemetryRepository) GetLatestByDeviceSN(ctx context.Context, deviceSN string) (*domainTelemetry.Telemetry, error) {
	fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1h)
		|> filter(fn: (r) => r["_measurement"] == "telemetry")
		|> filter(fn: (r) => r["device_sn"] == "%s")
		|> last()
	`, r.influx.bucket, deviceSN)

	result, err := r.queryFlux(ctx, fluxQuery)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

func (r *telemetryRepository) GetAllLatest(ctx context.Context) ([]*domainTelemetry.Telemetry, error) {
	fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1h)
		|> filter(fn: (r) => r["_measurement"] == "telemetry")
		|> last()
	`, r.influx.bucket)

	return r.queryFlux(ctx, fluxQuery)
}

func (r *telemetryRepository) QueryHistory(ctx context.Context, deviceSN string, query *domainTelemetry.TelemetryQuery) (*domainTelemetry.TelemetryListResult, error) {
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Order == "" {
		query.Order = "_time"
	}
	if query.Sort == "" {
		query.Sort = "DESC"
	}

	startTime := time.Now().Add(-24 * time.Hour)
	if query.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, query.StartTime); err == nil {
			startTime = t
		}
	}
	endTime := time.Now()
	if query.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, query.EndTime); err == nil {
			endTime = t
		}
	}

	offset := (query.Page - 1) * query.Limit
	descStr := "true"
	if query.Sort == "ASC" {
		descStr = "false"
	}

	fluxQuery := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "telemetry")
		|> filter(fn: (r) => r["device_sn"] == "%s")
		|> limit(n: %d, offset: %d)
		|> sort(columns: ["%s"], desc: %s)
	`, r.influx.bucket, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339),
		deviceSN, query.Limit, offset, query.Order, descStr)

	records, err := r.queryFlux(ctx, fluxQuery)
	if err != nil {
		return nil, err
	}

	total := int64(len(records))
	totalPages := int(total) / query.Limit
	if int(total)%query.Limit > 0 {
		totalPages++
	}

	return &domainTelemetry.TelemetryListResult{
		Telemetry: records,
		Total:     total,
		Page:      query.Page,
		Limit:     query.Limit,
		TotalPage: totalPages,
	}, nil
}

func (r *telemetryRepository) queryFlux(ctx context.Context, fluxQuery string) ([]*domainTelemetry.Telemetry, error) {
	params := url.Values{}
	params.Set("org", r.influx.org)

	reqURL := r.influx.url + "/api/v2/query?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(fluxQuery))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+r.influx.token)
	req.Header.Set("Content-Type", "application/vnd.influxql; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := r.influx.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("influxdb query error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result influxQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []*domainTelemetry.Telemetry{}, nil
	}

	return r.parseQueryResult(result), nil
}

type influxQueryResult struct {
	StatementID int `json:"statement_id"`
	Series     []struct {
		Name   string            `json:"name"`
		Tags   map[string]string `json:"tags"`
		Columns []string         `json:"columns"`
		Values  [][]interface{}  `json:"values"`
	} `json:"series"`
}

func (r *telemetryRepository) parseQueryResult(result influxQueryResult) []*domainTelemetry.Telemetry {
	if len(result.Series) == 0 {
		return []*domainTelemetry.Telemetry{}
	}

	telemetryMap := make(map[string]*domainTelemetry.Telemetry)

	for _, series := range result.Series {
		if series.Name != "telemetry" {
			continue
		}
		deviceSN := series.Tags["device_sn"]
		deviceType := series.Tags["device_type"]
		locationName := series.Tags["location_name"]

		for _, row := range series.Values {
			rec := make(map[string]interface{})
			for i, col := range series.Columns {
				if i < len(row) {
					rec[col] = row[i]
				}
			}

			ts := time.Now()
			if v, ok := rec["_time"].(string); ok {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					ts = t
				}
			}

			key := deviceSN + ts.Format(time.RFC3339Nano)
			t := &domainTelemetry.Telemetry{
				DeviceSN:     deviceSN,
				DeviceType:   deviceType,
				LocationName: locationName,
				Timestamp:    ts,
			}

			if v, ok := rec["temperature"].(float64); ok {
				t.Temperature = v
			}
			if v, ok := rec["humidity"].(float64); ok {
				t.Humidity = v
			}
			if v, ok := rec["aqi"].(float64); ok {
				t.AirQualityIndex = int(v)
			}
			if v, ok := rec["pm25"].(float64); ok {
				t.PM25 = v
			}
			if v, ok := rec["pm10"].(float64); ok {
				t.PM10 = v
			}
			if v, ok := rec["co2"].(float64); ok {
				t.CO2 = int(v)
			}
			if v, ok := rec["voc"].(float64); ok {
				t.VOC = v
			}

			telemetryMap[key] = t
		}
	}

	result_slice := make([]*domainTelemetry.Telemetry, 0, len(telemetryMap))
	for _, t := range telemetryMap {
		result_slice = append(result_slice, t)
	}
	return result_slice
}

// escapeTag escapes InfluxDB line protocol tag special characters
func escapeTag(s string) string {
	var buf bytes.Buffer
	for _, ch := range s {
		switch ch {
		case ',', ' ', '=', '\n', '\r', '\t':
			buf.WriteByte('_')
		default:
			buf.WriteRune(ch)
		}
	}
	return buf.String()
}

// ============================================================
// NEW: GetLatestHealth — query health measurement with pivot
// Pattern dari DEVICE_STREAM_GUIDE.md Section 5
// ============================================================

const offlineThreshold = 30 * time.Second

func (r *telemetryRepository) GetLatestHealth(ctx context.Context, filter domainTelemetry.DeviceFilter) ([]domainTelemetry.HealthData, error) {
	fluxQuery := r.buildHealthQuery(filter)

	result, err := r.queryFluxGeneric(ctx, fluxQuery)
	if err != nil {
		return nil, err
	}

	return r.parseHealthResult(result)
}

func (r *telemetryRepository) buildHealthQuery(filter domainTelemetry.DeviceFilter) string {
	projectFilter := ""
	if filter.Project != "" {
		projectFilter = fmt.Sprintf(`|> filter(fn: (r) => r["project"] == "%s")`, filter.Project)
	}
	deviceFilter := ""
	if filter.DeviceSN != "" {
		deviceFilter = fmt.Sprintf(`|> filter(fn: (r) => r["device_sn"] == "%s")`, filter.DeviceSN)
	}

	return fmt.Sprintf(`
from(bucket: "%s")
    |> range(start: -1m)
    |> filter(fn: (r) => r["_measurement"] == "health")
    %s
    %s
    |> group(columns: ["device_sn", "type", "project", "gateway_sn", "model"])
    |> last()
    |> pivot(
        rowKey:      ["device_sn", "type", "project", "gateway_sn", "model"],
        columnKey:   ["_field"],
        valueColumn: "_value"
    )
    |> group()
    |> yield(name: "health")
`, r.influx.bucket, projectFilter, deviceFilter)
}

func (r *telemetryRepository) parseHealthResult(result influxQueryResult) ([]domainTelemetry.HealthData, error) {
	var list []domainTelemetry.HealthData

	for _, series := range result.Series {
		if series.Name != "health" {
			continue
		}

		for _, row := range series.Values {
			rec := make(map[string]interface{})
			for i, col := range series.Columns {
				if i < len(row) {
					rec[col] = row[i]
				}
			}

			lastSeen := time.Now()
			if v, ok := rec["_time"].(string); ok {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					lastSeen = t
				}
			}

			status := "online"
			if time.Since(lastSeen) > offlineThreshold {
				status = "offline"
			}

			list = append(list, domainTelemetry.HealthData{
				DeviceSN:    toString(rec["device_sn"]),
				GatewaySN:   toString(rec["gateway_sn"]),
				Project:     toString(rec["project"]),
				Type:        toString(rec["type"]),
				Model:       toString(rec["model"]),
				Status:      status,
				Uptime:      toFloat64(rec["uptime"]),
				Temp:        toFloat64(rec["temp"]),
				Hum:         toFloat64(rec["hum"]),
				RSSI:        toFloat64(rec["rssi"]),
				ResetReason: toString(rec["reset_reason"]),
				LastSeen:    lastSeen,
			})
		}
	}

	return list, nil
}

// ============================================================
// NEW: GetLatestTelemetry — query telemetry measurement with pivot
// Pattern dari DEVICE_STREAM_GUIDE.md Section 5
// ============================================================

func (r *telemetryRepository) GetLatestTelemetry(ctx context.Context, filter domainTelemetry.DeviceFilter) (map[string]domainTelemetry.TelemetryData, error) {
	fluxQuery := r.buildLatestTelemetryQuery(filter)

	result, err := r.queryFluxGeneric(ctx, fluxQuery)
	if err != nil {
		return nil, err
	}

	return r.parseTelemetryLatest(result)
}

func (r *telemetryRepository) buildLatestTelemetryQuery(filter domainTelemetry.DeviceFilter) string {
	projectFilter := ""
	if filter.Project != "" {
		projectFilter = fmt.Sprintf(`|> filter(fn: (r) => r["project"] == "%s")`, filter.Project)
	}
	deviceFilter := ""
	if filter.DeviceSN != "" {
		deviceFilter = fmt.Sprintf(`|> filter(fn: (r) => r["device_sn"] == "%s")`, filter.DeviceSN)
	}

	return fmt.Sprintf(`
from(bucket: "%s")
    |> range(start: -1m)
    |> filter(fn: (r) => r["_measurement"] == "telemetry")
    %s
    %s
    |> group(columns: ["device_sn", "type", "project", "gateway_sn", "_field"])
    |> last()
    |> pivot(
        rowKey:      ["device_sn", "type", "project", "gateway_sn"],
        columnKey:   ["_field"],
        valueColumn: "_value"
    )
    |> group()
    |> yield(name: "telemetry")
`, r.influx.bucket, projectFilter, deviceFilter)
}

func (r *telemetryRepository) parseTelemetryLatest(result influxQueryResult) (map[string]domainTelemetry.TelemetryData, error) {
	skipCols := map[string]bool{
		"_time": true, "_start": true, "_stop": true,
		"_measurement": true, "device_sn": true,
		"gateway_sn": true, "project": true, "type": true,
		"result": true, "table": true,
	}

	telemetryMap := make(map[string]domainTelemetry.TelemetryData)

	for _, series := range result.Series {
		if series.Name != "telemetry" {
			continue
		}

		for _, row := range series.Values {
			rec := make(map[string]interface{})
			for i, col := range series.Columns {
				if i < len(row) {
					rec[col] = row[i]
				}
			}

			sn := toString(rec["device_sn"])
			if sn == "" {
				continue
			}

			if telemetryMap[sn] == nil {
				telemetryMap[sn] = make(domainTelemetry.TelemetryData)
			}

			for k, v := range rec {
				if skipCols[k] || v == nil {
					continue
				}
				telemetryMap[sn][k] = v
			}
		}
	}

	return telemetryMap, nil
}

// ============================================================
// NEW: GetTelemetryHistory — query history with time range
// Pattern dari DEVICE_STREAM_GUIDE.md Section 5
// ============================================================

func (r *telemetryRepository) GetTelemetryHistory(ctx context.Context, filter domainTelemetry.HistoryFilter) ([]domainTelemetry.TelemetryRecord, error) {
	fluxQuery := r.buildHistoryQuery(filter)

	result, err := r.queryFluxGeneric(ctx, fluxQuery)
	if err != nil {
		return nil, err
	}

	return r.parseTelemetryHistory(result)
}

func (r *telemetryRepository) buildHistoryQuery(filter domainTelemetry.HistoryFilter) string {
	projectFilter := ""
	if filter.Project != "" {
		projectFilter = fmt.Sprintf(`|> filter(fn: (r) => r["project"] == "%s")`, filter.Project)
	}

	aggregation := ""
	if filter.Window != "" {
		aggregation = fmt.Sprintf(`|> aggregateWindow(every: %s, fn: mean, createEmpty: false)`, filter.Window)
	}

	return fmt.Sprintf(`
from(bucket: "%s")
    |> range(start: %s, stop: %s)
    |> filter(fn: (r) => r["_measurement"] == "telemetry")
    |> filter(fn: (r) => r["device_sn"] == "%s")
    %s
    %s
    |> pivot(
        rowKey:      ["_time", "device_sn"],
        columnKey:   ["_field"],
        valueColumn: "_value"
    )
    |> sort(columns: ["_time"], desc: false)
    |> yield(name: "history")
`,
		r.influx.bucket,
		filter.TimeRange.Start.UTC().Format(time.RFC3339),
		filter.TimeRange.Stop.UTC().Format(time.RFC3339),
		filter.DeviceSN,
		projectFilter,
		aggregation)
}

func (r *telemetryRepository) parseTelemetryHistory(result influxQueryResult) ([]domainTelemetry.TelemetryRecord, error) {
	skipCols := map[string]bool{
		"_start": true, "_stop": true, "_measurement": true,
		"device_sn": true, "gateway_sn": true, "project": true,
		"type": true, "result": true, "table": true,
	}

	var records []domainTelemetry.TelemetryRecord

	for _, series := range result.Series {
		if series.Name != "history" {
			continue
		}

		for _, row := range series.Values {
			rec := make(map[string]interface{})
			for i, col := range series.Columns {
				if i < len(row) {
					rec[col] = row[i]
				}
			}

			fields := make(map[string]interface{})
			var timestamp time.Time

			for k, v := range rec {
				if k == "_time" {
					if ts, ok := v.(string); ok {
						if t, err := time.Parse(time.RFC3339, ts); err == nil {
							timestamp = t
						}
					}
					continue
				}
				if skipCols[k] || v == nil {
					continue
				}
				fields[k] = v
			}

			records = append(records, domainTelemetry.TelemetryRecord{
				Timestamp: timestamp,
				Fields:   fields,
			})
		}
	}

	return records, nil
}

// ============================================================
// Helper methods
// ============================================================

// queryFluxGeneric — generic Flux query returning raw JSON result
func (r *telemetryRepository) queryFluxGeneric(ctx context.Context, fluxQuery string) (influxQueryResult, error) {
	params := url.Values{}
	params.Set("org", r.influx.org)

	reqURL := r.influx.url + "/api/v2/query?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(fluxQuery))
	if err != nil {
		return influxQueryResult{}, err
	}
	req.Header.Set("Authorization", "Token "+r.influx.token)
	req.Header.Set("Content-Type", "application/vnd.influxql; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	resp, err := r.influx.httpClient.Do(req)
	if err != nil {
		return influxQueryResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return influxQueryResult{}, fmt.Errorf("influxdb query error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result influxQueryResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return influxQueryResult{}, err
	}

	return result, nil
}

// toFloat64 converts interface{} to float64 safely
func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case int:
		return float64(val)
	}
	return 0
}

// toString converts interface{} to string safely
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
