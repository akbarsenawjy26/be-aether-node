package telemetry

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	domainTelemetry "aether-node/internal/domain/telemetry"
	"aether-node/internal/circuitbreaker"
	"aether-node/internal/metrics"
)

// InfluxDB HTTP client — direct API calls, no SDK dependency
type influxHTTPClient struct {
	url    string
	token  string
	org    string
	bucket string
	httpClient *http.Client
	cb *circuitbreaker.CircuitBreaker
}

func newInfluxHTTPClient(url, token, org, bucket string) *influxHTTPClient {
	// Create HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,              // max idle connections per host
		MaxIdleConnsPerHost: 10,               // max idle per host (InfluxDB)
		IdleConnTimeout:     90 * time.Second, // how long idle connections live
		DisableKeepAlives:   false,
	}

	return &influxHTTPClient{
		url:    url,
		token:  token,
		org:    org,
		bucket: bucket,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		cb: circuitbreaker.New(circuitbreaker.Config{
			Name:                 "influxdb",
			FailureThreshold:     5,
			SuccessThreshold:     2,
			Timeout:              30 * time.Second,
			MaxConcurrentRequests: 10,
		}),
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
	req.Header.Set("Content-Type", "application/vnd.flux")
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

	start := time.Now()
	result, err := r.queryFluxGeneric(ctx, fluxQuery, false)
	duration := time.Since(start).Seconds()

	metrics.RecordInfluxDBQuery("get_latest_health", duration, err == nil)

	if err != nil {
		return nil, err
	}

	return r.parseHealthResult(result)
}

// SSE Bypass: same as GetLatestHealth but skips circuit breaker for long-lived SSE connections
func (r *telemetryRepository) GetLatestHealthSSE(ctx context.Context, filter domainTelemetry.DeviceFilter) ([]domainTelemetry.HealthData, error) {
	fluxQuery := r.buildHealthQuery(filter)

	start := time.Now()
	result, err := r.queryFluxGeneric(ctx, fluxQuery, true) // bypassCB=true
	duration := time.Since(start).Seconds()

	metrics.RecordInfluxDBQuery("get_latest_health_sse", duration, err == nil)

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
    |> group(columns: ["device_sn", "type", "project", "gateway_sn", "model", "_field"])
    |> last()
`, r.influx.bucket, projectFilter, deviceFilter)
}

func (r *telemetryRepository) parseHealthResult(result influxQueryResult) ([]domainTelemetry.HealthData, error) {
	// Group by device_sn - collect all field values for each device
	deviceMap := make(map[string]*domainTelemetry.HealthData)

	for _, series := range result.Series {
		// Skip non-health series
		if len(series.Columns) == 0 || len(series.Values) == 0 {
			continue
		}

		// Find column indices
		colIdx := make(map[string]int)
		for i, col := range series.Columns {
			colIdx[col] = i
		}

		for _, row := range series.Values {
			// Get device_sn (required)
			deviceSN := ""
			if i, ok := colIdx["device_sn"]; ok && i < len(row) {
				deviceSN = toString(row[i])
			}
			if deviceSN == "" {
				continue
			}

			// Get or create device entry
			if _, exists := deviceMap[deviceSN]; !exists {
				deviceMap[deviceSN] = &domainTelemetry.HealthData{
					DeviceSN: deviceSN,
				}
			}
			dev := deviceMap[deviceSN]

			// Extract tags/attributes
			if i, ok := colIdx["gateway_sn"]; ok && i < len(row) {
				dev.GatewaySN = toString(row[i])
			}
			if i, ok := colIdx["project"]; ok && i < len(row) {
				dev.Project = toString(row[i])
			}
			if i, ok := colIdx["type"]; ok && i < len(row) {
				dev.Type = toString(row[i])
			}
			if i, ok := colIdx["model"]; ok && i < len(row) {
				dev.Model = toString(row[i])
			}

			// Get _field and _value
			fieldName := ""
			var fieldValue interface{}
			if i, ok := colIdx["_field"]; ok && i < len(row) {
				fieldName = toString(row[i])
			}
			if i, ok := colIdx["_value"]; ok && i < len(row) {
				fieldValue = row[i]
			}

			// Parse field value based on field name
			switch fieldName {
			case "uptime":
				dev.Uptime = toFloat64(fieldValue)
			case "temp":
				dev.Temp = toFloat64(fieldValue)
			case "hum":
				dev.Hum = toFloat64(fieldValue)
			case "rssi":
				dev.RSSI = toFloat64(fieldValue)
			case "reset_reason":
				dev.ResetReason = toString(fieldValue)
			}

			// Get last_seen from _time
			if i, ok := colIdx["_time"]; ok && i < len(row) {
				if ts, ok := row[i].(string); ok {
					if t, err := time.Parse(time.RFC3339, ts); err == nil {
						if t.After(dev.LastSeen) {
							dev.LastSeen = t
						}
					}
				}
			}
		}
	}

	// Convert to slice
	list := make([]domainTelemetry.HealthData, 0, len(deviceMap))
	now := time.Now()
	for _, dev := range deviceMap {
		// Set status based on last_seen
		if dev.LastSeen.IsZero() {
			dev.LastSeen = now
		}
		if now.Sub(dev.LastSeen) > offlineThreshold {
			dev.Status = "offline"
		} else {
			dev.Status = "online"
		}
		list = append(list, *dev)
	}

	return list, nil
}

// ============================================================
// NEW: GetLatestTelemetry — query telemetry measurement with pivot
// Pattern dari DEVICE_STREAM_GUIDE.md Section 5
// ============================================================

func (r *telemetryRepository) GetLatestTelemetry(ctx context.Context, filter domainTelemetry.DeviceFilter) (map[string]domainTelemetry.TelemetryData, error) {
	fluxQuery := r.buildLatestTelemetryQuery(filter)

	start := time.Now()
	result, err := r.queryFluxGeneric(ctx, fluxQuery, false)
	duration := time.Since(start).Seconds()

	metrics.RecordInfluxDBQuery("get_latest_telemetry", duration, err == nil)

	if err != nil {
		return nil, err
	}

	return r.parseTelemetryLatest(result)
}

// SSE Bypass: same as GetLatestTelemetry but skips circuit breaker for long-lived SSE connections
func (r *telemetryRepository) GetLatestTelemetrySSE(ctx context.Context, filter domainTelemetry.DeviceFilter) (map[string]domainTelemetry.TelemetryData, error) {
	fluxQuery := r.buildLatestTelemetryQuery(filter)

	start := time.Now()
	result, err := r.queryFluxGeneric(ctx, fluxQuery, true) // bypassCB=true
	duration := time.Since(start).Seconds()

	metrics.RecordInfluxDBQuery("get_latest_telemetry_sse", duration, err == nil)

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
`, r.influx.bucket, projectFilter, deviceFilter)
}

func (r *telemetryRepository) parseTelemetryLatest(result influxQueryResult) (map[string]domainTelemetry.TelemetryData, error) {
	telemetryMap := make(map[string]domainTelemetry.TelemetryData)

	for _, series := range result.Series {
		if len(series.Columns) == 0 || len(series.Values) == 0 {
			continue
		}

		// Find column indices
		colIdx := make(map[string]int)
		for i, col := range series.Columns {
			colIdx[col] = i
		}

		for _, row := range series.Values {
			// Get device_sn
			sn := ""
			if i, ok := colIdx["device_sn"]; ok && i < len(row) {
				sn = toString(row[i])
			}
			if sn == "" {
				continue
			}

			if _, exists := telemetryMap[sn]; !exists {
				telemetryMap[sn] = make(domainTelemetry.TelemetryData)
			}

			// Get _field and _value
			fieldName := ""
			if i, ok := colIdx["_field"]; ok && i < len(row) {
				fieldName = toString(row[i])
			}
			if i, ok := colIdx["_value"]; ok && i < len(row) && row[i] != nil {
				telemetryMap[sn][fieldName] = row[i]
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

	start := time.Now()
	result, err := r.queryFluxGeneric(ctx, fluxQuery, false)
	duration := time.Since(start).Seconds()

	metrics.RecordInfluxDBQuery("get_telemetry_history", duration, err == nil)

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
    |> group(columns: ["_time", "device_sn", "_field"])
    |> last()
    |> sort(columns: ["_time"], desc: false)
`,
		r.influx.bucket,
		filter.TimeRange.Start.UTC().Format(time.RFC3339),
		filter.TimeRange.Stop.UTC().Format(time.RFC3339),
		filter.DeviceSN,
		projectFilter,
		aggregation)
}

func (r *telemetryRepository) parseTelemetryHistory(result influxQueryResult) ([]domainTelemetry.TelemetryRecord, error) {
	// Group records by timestamp
	timestampMap := make(map[string]*domainTelemetry.TelemetryRecord)

	for _, series := range result.Series {
		if len(series.Columns) == 0 || len(series.Values) == 0 {
			continue
		}

		// Find column indices
		colIdx := make(map[string]int)
		for i, col := range series.Columns {
			colIdx[col] = i
		}

		for _, row := range series.Values {
			// Get _time
			timestamp := time.Time{}
			if i, ok := colIdx["_time"]; ok && i < len(row) {
				if ts, ok := row[i].(string); ok {
					if t, err := time.Parse(time.RFC3339, ts); err == nil {
						timestamp = t
					}
				}
			}

			tsKey := timestamp.Format(time.RFC3339Nano)
			if _, exists := timestampMap[tsKey]; !exists {
				timestampMap[tsKey] = &domainTelemetry.TelemetryRecord{
					Timestamp: timestamp,
					Fields:   make(map[string]interface{}),
				}
			}

			// Get _field and _value
			fieldName := ""
			var fieldValue interface{}
			if i, ok := colIdx["_field"]; ok && i < len(row) {
				fieldName = toString(row[i])
			}
			if i, ok := colIdx["_value"]; ok && i < len(row) && row[i] != nil {
				fieldValue = row[i]
			}

			if fieldName != "" {
				timestampMap[tsKey].Fields[fieldName] = fieldValue
			}
		}
	}

	// Convert to slice and sort by timestamp
	records := make([]domainTelemetry.TelemetryRecord, 0, len(timestampMap))
	for _, rec := range timestampMap {
		records = append(records, *rec)
	}

	// Sort by timestamp ascending
	for i := 0; i < len(records)-1; i++ {
		for j := i + 1; j < len(records); j++ {
			if records[j].Timestamp.Before(records[i].Timestamp) {
				records[i], records[j] = records[j], records[i]
			}
		}
	}

	return records, nil
}

// ============================================================
// Helper methods
// ============================================================

// queryFluxGeneric — generic Flux query with circuit breaker + connection pooling
// Set bypassCB=true to skip circuit breaker (for long-lived SSE connections)
func (r *telemetryRepository) queryFluxGeneric(ctx context.Context, fluxQuery string, bypassCB bool) (influxQueryResult, error) {
	var result influxQueryResult
	var queryErr error

	// Define the actual query execution
	executeQuery := func() error {
		params := url.Values{}
		params.Set("org", r.influx.org)

		reqURL := r.influx.url + "/api/v2/query?" + params.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBufferString(fluxQuery))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Token "+r.influx.token)
		req.Header.Set("Content-Type", "application/vnd.flux")
		req.Header.Set("Accept", "application/json")

		resp, err := r.influx.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("influxdb query error: status=%d body=%s", resp.StatusCode, string(body))
		}

		// Parse CSV response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		result, err = parseFluxCSV(string(body))
		if err != nil {
			queryErr = err
		}

		return queryErr
	}

	var err error
	if bypassCB {
		// SSE/bypass: execute directly without circuit breaker
		err = executeQuery()
	} else {
		// Normal: execute through circuit breaker
		err = r.influx.cb.Execute(ctx, func(ctx context.Context) error {
			return executeQuery()
		})
	}

	if err != nil {
		return influxQueryResult{}, err
	}

	return result, nil
}

// parseFluxCSV parses InfluxDB 2.x Flux CSV response into influxQueryResult
// CSV format: ,result,table,_start,_stop,_time,_value,_field,_measurement,tags...
// Multiple tables (results) separated by empty lines
func parseFluxCSV(csvData string) (influxQueryResult, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	reader.FieldsPerRecord = -1 // Allow varying fields
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return influxQueryResult{}, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return influxQueryResult{}, nil
	}

	// Parse header row - first column is empty (result index)
	header := records[0]
	// header[0] is empty string "", columns start from index 1
	// columns: "", "result", "table", "_start", "_stop", "_time", "_value", "_field", "_measurement", ...

	// Build column index map
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	result := influxQueryResult{
		Series: []struct {
			Name   string            `json:"name"`
			Tags   map[string]string `json:"tags"`
			Columns []string         `json:"columns"`
			Values  [][]interface{}  `json:"values"`
		}{},
	}

	// Group rows by result+table
	tableMap := make(map[string]int) // key: "result,table" -> series index

	for rowIdx := 1; rowIdx < len(records); rowIdx++ {
		row := records[rowIdx]
		if len(row) == 0 {
			continue // skip empty lines
		}

		// Skip rows that don't have enough columns
		if len(row) < 5 {
			continue
		}

		// Parse result and table number
		resultName := ""
		tableNum := 0
		if idx, ok := colIndex["result"]; ok && idx < len(row) {
			resultName = row[idx]
		}
		if idx, ok := colIndex["table"]; ok && idx < len(row) {
			if v, err := strconv.Atoi(row[idx]); err == nil {
				tableNum = v
			}
		}

		key := fmt.Sprintf("%s,%d", resultName, tableNum)
		seriesIdx, exists := tableMap[key]
		if !exists {
			seriesIdx = len(result.Series)
			tableMap[key] = seriesIdx

			// Get measurement name
			measurementName := "unknown"
			if idx, ok := colIndex["_measurement"]; ok && idx < len(row) {
				measurementName = row[idx]
			}

			// Extract tag columns (columns that start with specific tags or are in the tag set)
			tags := make(map[string]string)
			tagColumns := []string{"device_sn", "project", "type", "gateway_sn", "model"}
			for _, tagCol := range tagColumns {
				if idx, ok := colIndex[tagCol]; ok && idx < len(row) && row[idx] != "" {
					tags[tagCol] = row[idx]
				}
			}

			// Extract column names (all columns except internal _ columns and result/table)
			columns := []string{}
			for i, col := range header {
				if i == 0 || col == "result" || col == "table" || col == "_start" || col == "_stop" || col == "_time" {
					continue
				}
				columns = append(columns, col)
			}

			result.Series = append(result.Series, struct {
				Name   string            `json:"name"`
				Tags   map[string]string `json:"tags"`
				Columns []string         `json:"columns"`
				Values  [][]interface{}  `json:"values"`
			}{
				Name:    measurementName,
				Tags:    tags,
				Columns: columns,
				Values:  [][]interface{}{},
			})
		}

		// Extract values for this row
		values := []interface{}{}
		for i, col := range header {
			if i == 0 || col == "result" || col == "table" || col == "_start" || col == "_stop" || col == "_time" {
				continue
			}
			if i < len(row) {
				values = append(values, parseValue(row[i]))
			} else {
				values = append(values, nil)
			}
		}

		result.Series[seriesIdx].Values = append(result.Series[seriesIdx].Values, values)
	}

	return result, nil
}

// parseValue converts string to appropriate type
func parseValue(s string) interface{} {
	if s == "" {
		return nil
	}
	// Try int
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	// Try float
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	// Return as string
	return s
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
