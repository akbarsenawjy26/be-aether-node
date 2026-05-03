# Device Stream & History — Implementation Guide
> Pedoman update/rewrite code existing untuk fitur SSE realtime dan history telemetry.
> Stack: Go · Echo · InfluxDB · MQTT (write only via MQTT, semua endpoint ini read-only)

---

## Daftar Isi
1. [Struktur Folder](#1-struktur-folder)
2. [Endpoints](#2-endpoints)
3. [Data Flow](#3-data-flow)
4. [InfluxDB — Skema Data](#4-influxdb--skema-data)
5. [InfluxDB — Flux Queries](#5-influxdb--flux-queries)
6. [Domain Layer](#6-domain-layer)
7. [Repository Layer](#7-repository-layer)
8. [Service Layer](#8-service-layer)
9. [Handler Layer](#9-handler-layer)
10. [Wire Up (main.go)](#10-wire-up-maingo)
11. [Payload Contoh](#11-payload-contoh)
12. [Catatan Penting](#12-catatan-penting)

---

## 1. Struktur Folder

```
internal/
└── device/
    ├── domain/
    │   ├── device.go       ← entities & filters
    │   └── repository.go   ← interface DeviceRepository
    ├── repository/
    │   └── influx_device_repository.go  ← implementasi InfluxDB (read only)
    ├── service/
    │   └── device_service.go  ← business logic
    └── handler/
        └── device_handler.go  ← Echo handler + SSE
```

---

## 2. Endpoints

| Method | Path | Tipe | Keterangan |
|--------|------|------|------------|
| `GET` | `/stream` | SSE | Semua device, grouped by `type` |
| `GET` | `/stream/:device_sn` | SSE | Per device, realtime health + telemetry |
| `POST` | `/history/telemetry/:device_sn` | REST | Time-series telemetry per device |

> **Semua endpoint read-only.** Write ke InfluxDB hanya via MQTT pipeline terpisah.

---

## 3. Data Flow

```
Client (EventSource / fetch)
    │
    ▼
Handler (Echo)
    │  SSE: polling setiap 5 detik
    │  History: single request
    ▼
Service
    │  GetAllDevices     → GetLatestHealth + GetLatestTelemetry → merge by device_sn
    │  GetDevice         → GetLatestHealth + GetLatestTelemetry → merge single device
    │  GetTelemetryHistory → GetTelemetryHistory (time-series)
    ▼
Repository (InfluxDB)
    │  Query 1: measurement "health"    → pivot by _field
    │  Query 2: measurement "telemetry" → pivot by _field
    ▼
InfluxDB (aether_dev_v2)
```

---

## 4. InfluxDB — Skema Data

### Measurement: `health`
| Key | Jenis | Nilai contoh |
|-----|-------|--------------|
| `project` | tag | `"project-a"` |
| `type` | tag | `"basic-model"`, `"aqi-model"` |
| `gateway_sn` | tag | `"GW001"` |
| `device_sn` | tag | `"SN001"` |
| `model` | tag | `"v2"` |
| `uptime` | field | `3600` |
| `temp` | field | `28.5` |
| `hum` | field | `65.2` |
| `rssi` | field | `-72` |
| `reset_reason` | field | `"power_on"` |

### Measurement: `telemetry`
| Key | Jenis | Nilai contoh |
|-----|-------|--------------|
| `project` | tag | `"project-a"` |
| `type` | tag | `"basic-model"` |
| `gateway_sn` | tag | `"GW001"` |
| `device_sn` | tag | `"SN001"` |
| *(fields dinamis)* | field | `voltage`, `current`, dll — berbeda per `type` |

---

## 5. InfluxDB — Flux Queries

### Query: Latest Health (semua device)
```flux
from(bucket: "aether_dev_v2")
    |> range(start: -1m)
    |> filter(fn: (r) => r._measurement == "health")
    // optional: |> filter(fn: (r) => r.project == "project-a")
    // optional: |> filter(fn: (r) => r.device_sn == "SN001")
    |> group(columns: ["device_sn", "type", "project", "gateway_sn", "model"])
    |> last()
    |> pivot(
        rowKey:      ["device_sn", "type", "project", "gateway_sn", "model"],
        columnKey:   ["_field"],
        valueColumn: "_value"
    )
    |> group()
    |> yield(name: "health")
```

### Query: Latest Telemetry (semua device)
```flux
from(bucket: "aether_dev_v2")
    |> range(start: -1m)
    |> filter(fn: (r) => r._measurement == "telemetry")
    // optional: |> filter(fn: (r) => r.project == "project-a")
    // optional: |> filter(fn: (r) => r.device_sn == "SN001")
    |> group(columns: ["device_sn", "type", "project", "gateway_sn", "_field"])
    |> last()
    |> pivot(
        rowKey:      ["device_sn", "type", "project", "gateway_sn"],
        columnKey:   ["_field"],
        valueColumn: "_value"
    )
    |> group()
    |> yield(name: "telemetry")
```

### Query: Telemetry History (per device, dengan time range)
```flux
from(bucket: "aether_dev_v2")
    |> range(start: 2024-01-01T00:00:00Z, stop: 2024-01-31T23:59:59Z)
    |> filter(fn: (r) => r._measurement == "telemetry")
    |> filter(fn: (r) => r.device_sn == "SN001")
    // optional: |> filter(fn: (r) => r.project == "project-a")
    // optional aggregasi: |> aggregateWindow(every: 5m, fn: mean, createEmpty: false)
    |> pivot(
        rowKey:      ["_time", "device_sn"],
        columnKey:   ["_field"],
        valueColumn: "_value"
    )
    |> sort(columns: ["_time"], desc: false)
    |> yield(name: "history")
```

> **Catatan Flux:**
> - `group()` sebelum `last()` wajib agar `last()` bekerja per device, bukan global
> - `rowKey` harus mencakup semua tag yang jadi identifier unik row
> - `|> group()` tanpa argumen di akhir = ungroup, wajib sebelum `join()` klasik
> - Format waktu: RFC3339 (`2024-01-01T00:00:00Z`), selalu UTC
> - Threshold offline: `time.Since(lastSeen) > 30 * time.Second`

---

## 6. Domain Layer

```go
// domain/device.go
package domain

import "time"

type HealthData struct {
    DeviceSN    string    `json:"device_sn"`
    GatewaySN   string    `json:"gateway_sn"`
    Project     string    `json:"project"`
    Type        string    `json:"type"`
    Model       string    `json:"model"`
    Status      string    `json:"status"`      // "online" | "offline"
    Uptime      float64   `json:"uptime"`
    Temp        float64   `json:"temp"`
    Hum         float64   `json:"hum"`
    RSSI        float64   `json:"rssi"`
    ResetReason string    `json:"reset_reason"`
    LastSeen    time.Time `json:"last_seen"`
}

// TelemetryData: fields dinamis per device type
type TelemetryData map[string]interface{}

// TelemetryRecord: satu titik waktu untuk history
type TelemetryRecord struct {
    Timestamp time.Time              `json:"timestamp"`
    Fields    map[string]interface{} `json:"fields"`
}

// DeviceEntry: satu device lengkap (SSE payload item)
type DeviceEntry struct {
    Health    HealthData    `json:"health"`
    Telemetry TelemetryData `json:"telemetry"`
}

// DevicePayload: SSE dashboard payload
// {"basic-model": [...], "aqi-model": [...]}
type DevicePayload map[string][]DeviceEntry

type QueryTimeRange struct {
    Start time.Time
    Stop  time.Time
}

// DeviceFilter: untuk query latest (SSE)
type DeviceFilter struct {
    Project  string
    DeviceSN string // kosong = semua device
}

// HistoryFilter: untuk query history
type HistoryFilter struct {
    Project   string
    DeviceSN  string
    TimeRange QueryTimeRange
    Window    string // "1m", "5m", "1h" — kosong = raw data
}
```

```go
// domain/repository.go
package domain

import "context"

type DeviceRepository interface {
    GetLatestHealth(ctx context.Context, bucket string, filter DeviceFilter) ([]HealthData, error)
    GetLatestTelemetry(ctx context.Context, bucket string, filter DeviceFilter) (map[string]TelemetryData, error)
    GetTelemetryHistory(ctx context.Context, bucket string, filter HistoryFilter) ([]TelemetryRecord, error)
}
```

---

## 7. Repository Layer

```go
// repository/influx_device_repository.go
package repository

import (
    "context"
    "fmt"
    "time"

    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/influxdata/influxdb-client-go/v2/api"

    "yourapp/device/domain"
)

const offlineThreshold = 30 * time.Second

type deviceRepository struct {
    queryAPI api.QueryAPI
}

func NewDeviceRepository(client influxdb2.Client, org string) domain.DeviceRepository {
    return &deviceRepository{
        queryAPI: client.QueryAPI(org),
    }
}

// ---------------------------------------------------------
// GetLatestHealth
// ---------------------------------------------------------

func (r *deviceRepository) GetLatestHealth(
    ctx context.Context, bucket string, filter domain.DeviceFilter,
) ([]domain.HealthData, error) {
    result, err := r.queryAPI.Query(ctx, r.buildHealthQuery(bucket, filter))
    if err != nil {
        return nil, fmt.Errorf("influx health query: %w", err)
    }
    defer result.Close()
    return r.parseHealthResult(result)
}

func (r *deviceRepository) buildHealthQuery(bucket string, filter domain.DeviceFilter) string {
    deviceFilter, projectFilter := buildFilters(filter.DeviceSN, filter.Project)
    return fmt.Sprintf(`
        from(bucket: "%s")
            |> range(start: -1m)
            |> filter(fn: (r) => r._measurement == "health")
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
    `, bucket, projectFilter, deviceFilter)
}

func (r *deviceRepository) parseHealthResult(result *api.QueryTableResult) ([]domain.HealthData, error) {
    var list []domain.HealthData
    for result.Next() {
        rec := result.Record()
        vals := rec.Values()
        lastSeen := rec.Time()
        status := "online"
        if time.Since(lastSeen) > offlineThreshold {
            status = "offline"
        }
        list = append(list, domain.HealthData{
            DeviceSN:    toString(rec.ValueByKey("device_sn")),
            GatewaySN:   toString(rec.ValueByKey("gateway_sn")),
            Project:     toString(rec.ValueByKey("project")),
            Type:        toString(rec.ValueByKey("type")),
            Model:       toString(rec.ValueByKey("model")),
            Status:      status,
            LastSeen:    lastSeen,
            Uptime:      toFloat64(vals["uptime"]),
            Temp:        toFloat64(vals["temp"]),
            Hum:         toFloat64(vals["hum"]),
            RSSI:        toFloat64(vals["rssi"]),
            ResetReason: toString(vals["reset_reason"]),
        })
    }
    if err := result.Err(); err != nil {
        return nil, fmt.Errorf("parsing health result: %w", err)
    }
    return list, nil
}

// ---------------------------------------------------------
// GetLatestTelemetry
// ---------------------------------------------------------

func (r *deviceRepository) GetLatestTelemetry(
    ctx context.Context, bucket string, filter domain.DeviceFilter,
) (map[string]domain.TelemetryData, error) {
    result, err := r.queryAPI.Query(ctx, r.buildLatestTelemetryQuery(bucket, filter))
    if err != nil {
        return nil, fmt.Errorf("influx telemetry query: %w", err)
    }
    defer result.Close()
    return r.parseTelemetryLatest(result)
}

func (r *deviceRepository) buildLatestTelemetryQuery(bucket string, filter domain.DeviceFilter) string {
    deviceFilter, projectFilter := buildFilters(filter.DeviceSN, filter.Project)
    return fmt.Sprintf(`
        from(bucket: "%s")
            |> range(start: -1m)
            |> filter(fn: (r) => r._measurement == "telemetry")
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
    `, bucket, projectFilter, deviceFilter)
}

func (r *deviceRepository) parseTelemetryLatest(result *api.QueryTableResult) (map[string]domain.TelemetryData, error) {
    skipCols := map[string]bool{
        "_time": true, "_start": true, "_stop": true,
        "_measurement": true, "device_sn": true,
        "gateway_sn": true, "project": true, "type": true,
        "result": true, "table": true,
    }
    telemetryMap := make(map[string]domain.TelemetryData)
    for result.Next() {
        rec := result.Record()
        sn := toString(rec.ValueByKey("device_sn"))
        if sn == "" {
            continue
        }
        if telemetryMap[sn] == nil {
            telemetryMap[sn] = make(domain.TelemetryData)
        }
        for k, v := range rec.Values() {
            if skipCols[k] || v == nil {
                continue
            }
            telemetryMap[sn][k] = v
        }
    }
    if err := result.Err(); err != nil {
        return nil, fmt.Errorf("parsing telemetry latest: %w", err)
    }
    return telemetryMap, nil
}

// ---------------------------------------------------------
// GetTelemetryHistory
// ---------------------------------------------------------

func (r *deviceRepository) GetTelemetryHistory(
    ctx context.Context, bucket string, filter domain.HistoryFilter,
) ([]domain.TelemetryRecord, error) {
    result, err := r.queryAPI.Query(ctx, r.buildHistoryQuery(bucket, filter))
    if err != nil {
        return nil, fmt.Errorf("influx telemetry history query: %w", err)
    }
    defer result.Close()
    return r.parseTelemetryHistory(result)
}

func (r *deviceRepository) buildHistoryQuery(bucket string, filter domain.HistoryFilter) string {
    _, projectFilter := buildFilters("", filter.Project)
    aggregation := ""
    if filter.Window != "" {
        aggregation = fmt.Sprintf(
            `|> aggregateWindow(every: %s, fn: mean, createEmpty: false)`,
            filter.Window,
        )
    }
    return fmt.Sprintf(`
        from(bucket: "%s")
            |> range(start: %s, stop: %s)
            |> filter(fn: (r) => r._measurement == "telemetry")
            |> filter(fn: (r) => r.device_sn == "%s")
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
        bucket,
        filter.TimeRange.Start.UTC().Format(time.RFC3339),
        filter.TimeRange.Stop.UTC().Format(time.RFC3339),
        filter.DeviceSN,
        projectFilter,
        aggregation,
    )
}

func (r *deviceRepository) parseTelemetryHistory(result *api.QueryTableResult) ([]domain.TelemetryRecord, error) {
    skipCols := map[string]bool{
        "_start": true, "_stop": true, "_measurement": true,
        "device_sn": true, "gateway_sn": true, "project": true,
        "type": true, "result": true, "table": true,
    }
    var records []domain.TelemetryRecord
    for result.Next() {
        rec := result.Record()
        fields := make(map[string]interface{})
        for k, v := range rec.Values() {
            if k == "_time" || skipCols[k] || v == nil {
                continue
            }
            fields[k] = v
        }
        records = append(records, domain.TelemetryRecord{
            Timestamp: rec.Time(),
            Fields:    fields,
        })
    }
    if err := result.Err(); err != nil {
        return nil, fmt.Errorf("parsing telemetry history: %w", err)
    }
    return records, nil
}

// ---------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------

func buildFilters(deviceSN, project string) (deviceFilter, projectFilter string) {
    if deviceSN != "" {
        deviceFilter = fmt.Sprintf(`|> filter(fn: (r) => r.device_sn == "%s")`, deviceSN)
    }
    if project != "" {
        projectFilter = fmt.Sprintf(`|> filter(fn: (r) => r.project == "%s")`, project)
    }
    return
}

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
    }
    return 0
}

func toString(v interface{}) string {
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", v)
}
```

---

## 8. Service Layer

```go
// service/device_service.go
package service

import (
    "context"
    "fmt"

    "yourapp/device/domain"
)

type DeviceService interface {
    GetAllDevices(ctx context.Context, project string) (domain.DevicePayload, error)
    GetDevice(ctx context.Context, project, deviceSN string) (*domain.DeviceEntry, error)
    GetTelemetryHistory(ctx context.Context, filter domain.HistoryFilter) ([]domain.TelemetryRecord, error)
}

type deviceService struct {
    repo   domain.DeviceRepository
    bucket string
}

func NewDeviceService(repo domain.DeviceRepository, bucket string) DeviceService {
    return &deviceService{repo: repo, bucket: bucket}
}

func (s *deviceService) GetAllDevices(ctx context.Context, project string) (domain.DevicePayload, error) {
    filter := domain.DeviceFilter{Project: project}

    healthList, err := s.repo.GetLatestHealth(ctx, s.bucket, filter)
    if err != nil {
        return nil, fmt.Errorf("get all health: %w", err)
    }
    telemetryMap, err := s.repo.GetLatestTelemetry(ctx, s.bucket, filter)
    if err != nil {
        return nil, fmt.Errorf("get all telemetry: %w", err)
    }

    payload := make(domain.DevicePayload)
    for _, h := range healthList {
        t := telemetryMap[h.DeviceSN]
        if t == nil {
            t = domain.TelemetryData{}
        }
        payload[h.Type] = append(payload[h.Type], domain.DeviceEntry{Health: h, Telemetry: t})
    }
    return payload, nil
}

func (s *deviceService) GetDevice(ctx context.Context, project, deviceSN string) (*domain.DeviceEntry, error) {
    filter := domain.DeviceFilter{Project: project, DeviceSN: deviceSN}

    healthList, err := s.repo.GetLatestHealth(ctx, s.bucket, filter)
    if err != nil {
        return nil, fmt.Errorf("get device health [%s]: %w", deviceSN, err)
    }
    if len(healthList) == 0 {
        return nil, fmt.Errorf("device not found: %s", deviceSN)
    }

    telemetryMap, err := s.repo.GetLatestTelemetry(ctx, s.bucket, filter)
    if err != nil {
        return nil, fmt.Errorf("get device telemetry [%s]: %w", deviceSN, err)
    }

    t := telemetryMap[deviceSN]
    if t == nil {
        t = domain.TelemetryData{}
    }
    return &domain.DeviceEntry{Health: healthList[0], Telemetry: t}, nil
}

func (s *deviceService) GetTelemetryHistory(
    ctx context.Context, filter domain.HistoryFilter,
) ([]domain.TelemetryRecord, error) {
    records, err := s.repo.GetTelemetryHistory(ctx, s.bucket, filter)
    if err != nil {
        return nil, fmt.Errorf("get telemetry history [%s]: %w", filter.DeviceSN, err)
    }
    return records, nil
}
```

---

## 9. Handler Layer

```go
// handler/device_handler.go
package handler

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/labstack/echo/v4"

    "yourapp/device/domain"
    "yourapp/device/service"
)

const (
    sseInterval    = 5 * time.Second
    sseContentType = "text/event-stream"
)

type DeviceHandler struct {
    svc service.DeviceService
}

func NewDeviceHandler(svc service.DeviceService) *DeviceHandler {
    return &DeviceHandler{svc: svc}
}

// RegisterRoutes — panggil dengan echo.Group yang sudah ada prefix-nya
// Contoh: h.RegisterRoutes(e.Group("/api/v1/devices"))
func (h *DeviceHandler) RegisterRoutes(g *echo.Group) {
    g.GET("/stream", h.StreamAllDevices)
    g.GET("/stream/:device_sn", h.StreamDevice)
    g.POST("/history/telemetry/:device_sn", h.GetTelemetryHistory)
}

// GET /stream?project=project-a
func (h *DeviceHandler) StreamAllDevices(c echo.Context) error {
    project := c.QueryParam("project")
    setSSEHeaders(c)

    ticker := time.NewTicker(sseInterval)
    defer ticker.Stop()

    if err := h.sendAllDevices(c, project); err != nil {
        return err
    }
    for {
        select {
        case <-c.Request().Context().Done():
            return nil
        case <-ticker.C:
            if err := h.sendAllDevices(c, project); err != nil {
                return err
            }
        }
    }
}

func (h *DeviceHandler) sendAllDevices(c echo.Context, project string) error {
    payload, err := h.svc.GetAllDevices(c.Request().Context(), project)
    if err != nil {
        writeSSEError(c, err)
        return nil
    }
    return writeSSEData(c, payload)
}

// GET /stream/:device_sn?project=project-a
func (h *DeviceHandler) StreamDevice(c echo.Context) error {
    deviceSN := c.Param("device_sn")
    project := c.QueryParam("project")
    setSSEHeaders(c)

    ticker := time.NewTicker(sseInterval)
    defer ticker.Stop()

    if err := h.sendDevice(c, project, deviceSN); err != nil {
        return err
    }
    for {
        select {
        case <-c.Request().Context().Done():
            return nil
        case <-ticker.C:
            if err := h.sendDevice(c, project, deviceSN); err != nil {
                return err
            }
        }
    }
}

func (h *DeviceHandler) sendDevice(c echo.Context, project, deviceSN string) error {
    entry, err := h.svc.GetDevice(c.Request().Context(), project, deviceSN)
    if err != nil {
        writeSSEError(c, err)
        return nil
    }
    return writeSSEData(c, entry)
}

// POST /history/telemetry/:device_sn?project=project-a
// Body: { "start": "RFC3339", "stop": "RFC3339", "window": "5m" }
type historyRequest struct {
    Start  string `json:"start"`
    Stop   string `json:"stop"`
    Window string `json:"window"` // optional: "1m", "5m", "1h"
}

type historyResponse struct {
    DeviceSN string                   `json:"device_sn"`
    Data     []domain.TelemetryRecord `json:"data"`
}

func (h *DeviceHandler) GetTelemetryHistory(c echo.Context) error {
    deviceSN := c.Param("device_sn")
    project := c.QueryParam("project")

    var req historyRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
    }
    if req.Start == "" || req.Stop == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "'start' and 'stop' are required")
    }

    start, err := time.Parse(time.RFC3339, req.Start)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start': use RFC3339 (e.g. 2024-01-01T00:00:00Z)")
    }
    stop, err := time.Parse(time.RFC3339, req.Stop)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid 'stop': use RFC3339 (e.g. 2024-01-31T23:59:59Z)")
    }
    if stop.Before(start) {
        return echo.NewHTTPError(http.StatusBadRequest, "'stop' must be after 'start'")
    }

    records, err := h.svc.GetTelemetryHistory(c.Request().Context(), domain.HistoryFilter{
        Project:  project,
        DeviceSN: deviceSN,
        TimeRange: domain.QueryTimeRange{
            Start: start.UTC(),
            Stop:  stop.UTC(),
        },
        Window: req.Window,
    })
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    }

    if records == nil {
        records = []domain.TelemetryRecord{}
    }
    return c.JSON(http.StatusOK, historyResponse{DeviceSN: deviceSN, Data: records})
}

// ---------------------------------------------------------
// SSE helpers
// ---------------------------------------------------------

func setSSEHeaders(c echo.Context) {
    c.Response().Header().Set(echo.HeaderContentType, sseContentType)
    c.Response().Header().Set("Cache-Control", "no-cache")
    c.Response().Header().Set("Connection", "keep-alive")
    c.Response().Header().Set("X-Accel-Buffering", "no") // penting untuk nginx
    c.Response().WriteHeader(http.StatusOK)
}

func writeSSEData(c echo.Context, data interface{}) error {
    b, err := json.Marshal(data)
    if err != nil {
        return err
    }
    _, err = fmt.Fprintf(c.Response(), "data: %s\n\n", b)
    if err != nil {
        return err
    }
    c.Response().Flush()
    return nil
}

func writeSSEError(c echo.Context, err error) {
    fmt.Fprintf(c.Response(), "event: error\ndata: %s\n\n", err.Error())
    c.Response().Flush()
}
```

---

## 10. Wire Up (main.go)

```go
// Di main.go atau di wire/provider
import (
    influxdb2 "github.com/influxdata/influxdb-client-go/v2"
    "github.com/labstack/echo/v4"

    devicerepo    "yourapp/device/repository"
    devicesvc     "yourapp/device/service"
    devicehandler "yourapp/device/handler"
)

func main() {
    e := echo.New()

    influxClient := influxdb2.NewClient("http://localhost:8086", "your-token")
    defer influxClient.Close()

    repo := devicerepo.NewDeviceRepository(influxClient, "your-org")
    svc  := devicesvc.NewDeviceService(repo, "aether_dev_v2")
    h    := devicehandler.NewDeviceHandler(svc)

    h.RegisterRoutes(e.Group("/api/v1/devices"))

    e.Start(":8080")
}
```

---

## 11. Payload Contoh

### SSE Dashboard — `GET /stream`
```json
{
  "basic-model": [
    {
      "health": {
        "device_sn": "SN001",
        "gateway_sn": "GW001",
        "project": "project-a",
        "type": "basic-model",
        "model": "v2",
        "status": "online",
        "uptime": 3600,
        "temp": 28.5,
        "hum": 65.2,
        "rssi": -72,
        "reset_reason": "power_on",
        "last_seen": "2024-01-01T07:00:00+07:00"
      },
      "telemetry": {
        "voltage": 220.5,
        "current": 1.2
      }
    }
  ],
  "aqi-model": [
    {
      "health": { "device_sn": "SN002", "status": "offline", "..." : "..." },
      "telemetry": {}
    }
  ]
}
```

### SSE Per Device — `GET /stream/SN001`
```json
{
  "health": {
    "device_sn": "SN001",
    "status": "online",
    "temp": 28.5,
    "...": "..."
  },
  "telemetry": {
    "voltage": 220.5,
    "current": 1.2
  }
}
```

### History — `POST /history/telemetry/SN001`
Request:
```json
{
  "start":  "2024-01-01T00:00:00Z",
  "stop":   "2024-01-31T23:59:59Z",
  "window": "5m"
}
```
Response:
```json
{
  "device_sn": "SN001",
  "data": [
    { "timestamp": "2024-01-01T00:00:00Z", "fields": { "voltage": 220.5, "current": 1.2 } },
    { "timestamp": "2024-01-01T00:05:00Z", "fields": { "voltage": 221.0, "current": 1.1 } }
  ]
}
```

---

## 12. Catatan Penting

### Flux — Kesalahan Umum yang Harus Dihindari
| Masalah | Penyebab | Fix |
|---------|----------|-----|
| `pivot: specified column does not exist` | Kolom tidak ada setelah `last()` | Pastikan `group()` sebelum `last()` mencakup kolom yang dibutuhkan |
| `no streaming data` | Semua hasil di-assign ke variabel, tidak ada output | Tambahkan `\|> yield(name: "result")` di akhir |
| `union` memisahkan baris | Group key berbeda antar tabel | Gunakan `join()` bukan `union()` jika ingin merge per device |
| Syntax error di query | String literal menggantung (sisa komentar `//`) | Hapus baris yang tidak valid |
| `join()` error lintas bucket | Flux tidak support join antar bucket | Samakan bucket untuk semua subquery |

### SSE — Hal Penting
- Header `X-Accel-Buffering: no` wajib jika di balik nginx, agar data tidak di-buffer
- Kirim data pertama **langsung** sebelum masuk loop ticker — menghindari delay 5 detik di client
- Error di dalam SSE loop **tidak** memutus koneksi — gunakan `event: error` dan lanjutkan
- `context.Done()` digunakan untuk detect client disconnect dan cleanup goroutine

### Status Online/Offline
- Dihitung di repository saat parse result, bukan di InfluxDB
- Threshold default: `30 * time.Second` dari `rec.Time()` (timestamp data terakhir)
- Sesuaikan threshold dengan interval publish MQTT device

### Telemetry Fields
- Fields telemetry bersifat **dinamis** per `type` device — menggunakan `map[string]interface{}`
- Kolom metadata InfluxDB (`_time`, `_start`, `_stop`, `_measurement`, `result`, `table`) di-skip saat parsing
- Jika field `nil`, di-skip — tidak akan muncul di JSON response
