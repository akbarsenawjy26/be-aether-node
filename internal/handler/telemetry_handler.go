package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"aether-node/internal/domain/telemetry"
	"aether-node/internal/metrics"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

const (
	sseInterval    = 5 * time.Second
	sseContentType = "text/event-stream"
)

type TelemetryHandler struct {
	svc telemetry.TelemetryService
}

func NewTelemetryHandler(svc telemetry.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{svc: svc}
}

// RegisterRoutes registers telemetry routes to the given Echo group
// Routes:
//   - GET  /stream             → StreamAllDevices (SSE all devices grouped by type)
//   - GET  /stream/:device_sn  → StreamDevice (SSE per device)
//   - POST /history/:device_sn → GetTelemetryHistory (REST time-series)
func (h *TelemetryHandler) RegisterRoutes(g *echo.Group) {
	g.GET("/stream", h.StreamAllDevices)
	g.GET("/stream/:device_sn", h.StreamDevice)
	g.POST("/history/:device_sn", h.GetTelemetryHistory)
}

// ============================================================
// SSE: GET /stream?project=xxx
// Pattern dari DEVICE_STREAM_GUIDE.md Section 9
// ============================================================

// StreamAllDevices handles GET /stream - SSE for all devices grouped by type
func (h *TelemetryHandler) StreamAllDevices(c echo.Context) error {
	project := c.QueryParam("project")

	// Set SSE headers
	setSSEHeaders(c)

	// Record SSE connection open
	metrics.RecordSSEConnection(true)
	defer metrics.RecordSSEConnection(false)

	// Send first data immediately (avoid 5s delay on client)
	if err := h.sendAllDevices(c, project); err != nil {
		writeSSEError(c, err)
		c.Response().Flush()
	}

	// Start polling loop
	ticker := time.NewTicker(sseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case <-ticker.C:
			if err := h.sendAllDevices(c, project); err != nil {
				writeSSEError(c, err)
				c.Response().Flush()
				continue // Don't disconnect on error
			}
		}
	}
}

func (h *TelemetryHandler) sendAllDevices(c echo.Context, project string) error {
	payload, err := h.svc.StreamAllDevicesWithHealth(c.Request().Context(), project)
	if err != nil {
		return err
	}

	// Record SSE events sent (grouped by type)
	for deviceType := range payload {
		metrics.RecordSSEEvent(deviceType)
	}

	return writeSSEData(c, payload)
}

// ============================================================
// SSE: GET /stream/:device_sn?project=xxx
// Pattern dari DEVICE_STREAM_GUIDE.md Section 9
// ============================================================

// StreamDevice handles GET /stream/:device_sn - SSE for specific device
func (h *TelemetryHandler) StreamDevice(c echo.Context) error {
	deviceSN := c.Param("device_sn")
	project := c.QueryParam("project")

	if deviceSN == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "device_sn is required")
	}

	// Set SSE headers
	setSSEHeaders(c)

	// Record SSE connection open
	metrics.RecordSSEConnection(true)
	defer metrics.RecordSSEConnection(false)

	// Send first data immediately (avoid 5s delay on client)
	if err := h.sendDevice(c, project, deviceSN); err != nil {
		writeSSEError(c, err)
		c.Response().Flush()
	}

	// Start polling loop
	ticker := time.NewTicker(sseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case <-ticker.C:
			if err := h.sendDevice(c, project, deviceSN); err != nil {
				writeSSEError(c, err)
				c.Response().Flush()
				continue // Don't disconnect on error
			}
		}
	}
}

func (h *TelemetryHandler) sendDevice(c echo.Context, project, deviceSN string) error {
	entry, err := h.svc.StreamDeviceWithHealth(c.Request().Context(), project, deviceSN)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("device not found: %s", deviceSN)
	}

	// Record SSE event sent (by device type)
	if entry.Health.Type != "" {
		metrics.RecordSSEEvent(entry.Health.Type)
	}

	return writeSSEData(c, entry)
}

// ============================================================
// REST: POST /history/:device_sn?project=xxx
// Pattern dari DEVICE_STREAM_GUIDE.md Section 9
// ============================================================

// historyRequest represents the request body for history endpoint
type historyRequest struct {
	Start string `json:"start"` // RFC3339 format, e.g., "2024-01-01T00:00:00Z"
	Stop  string `json:"stop"`  // RFC3339 format
	Window string `json:"window"` // optional: "1m", "5m", "1h"
	Order string `json:"order"` // "desc" or "asc" (default desc)
	Sort  string `json:"sort"`  // field to sort by (default "timestamp")
	Page  int    `json:"page"`  // page number (default 1)
	Limit int    `json:"limit"` // items per page (default 100, max 1000)
}

// historyResponse represents the response for history endpoint
type historyResponse struct {
	DeviceSN string                   `json:"device_sn"`
	Data     []telemetry.TelemetryRecord `json:"data"`
}

// GetTelemetryHistory handles POST /history/:device_sn - REST time-series telemetry
func (h *TelemetryHandler) GetTelemetryHistory(c echo.Context) error {
	deviceSN := c.Param("device_sn")
	project := c.QueryParam("project")

	if deviceSN == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "device_sn is required")
	}

	var req historyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Start == "" || req.Stop == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "'start' and 'stop' are required (RFC3339 format)")
	}

	// Parse time
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'start': use RFC3339 format (e.g., 2024-01-01T00:00:00Z)")
	}

	stop, err := time.Parse(time.RFC3339, req.Stop)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid 'stop': use RFC3339 format (e.g., 2024-01-31T23:59:59Z)")
	}

	if stop.Before(start) {
		return echo.NewHTTPError(http.StatusBadRequest, "'stop' must be after 'start'")
	}

	// Set defaults for pagination and sorting
	if req.Order == "" {
		req.Order = "desc"
	}
	if req.Sort == "" {
		req.Sort = "timestamp"
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 1000 {
		req.Limit = 100
	}

	sortDesc := req.Order == "desc"

	// Query history
	filter := telemetry.HistoryFilter{
		Project:  project,
		DeviceSN: deviceSN,
		TimeRange: telemetry.QueryTimeRange{
			Start: start.UTC(),
			Stop:  stop.UTC(),
		},
		Window:   req.Window,
		SortDesc: sortDesc,
		Limit:    req.Limit,
	}

	records, err := h.svc.GetTelemetryHistory(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if records == nil {
		records = []telemetry.TelemetryRecord{}
	}

	return c.JSON(http.StatusOK, historyResponse{
		DeviceSN: deviceSN,
		Data:     records,
	})
}

// ============================================================
// SSE Helper Functions
// Pattern dari DEVICE_STREAM_GUIDE.md Section 9
// ============================================================

// setSSEHeaders sets required headers for SSE
func setSSEHeaders(c echo.Context) {
	c.Response().Header().Set(echo.HeaderContentType, sseContentType)
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no") // Critical for nginx
	c.Response().WriteHeader(http.StatusOK)
}

// writeSSEData sends SSE data event
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

// writeSSEError sends SSE error event (does not disconnect)
func writeSSEError(c echo.Context, err error) {
	fmt.Fprintf(c.Response(), "event: error\ndata: %s\n\n", err.Error())
}

// ============================================================
// Legacy Handlers (kept for backward compatibility)
// ============================================================

// StreamAllDevicesLegacy handles GET /stream-legacy - old channel-based SSE
func (h *TelemetryHandler) StreamAllDevicesLegacy(c echo.Context) error {
	ctx := c.Request().Context()

	c.Response().Header().Set("Content-Type", sseContentType)
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	telemetryChan, errChan := h.svc.StreamAllDevices(ctx)
	ticker := time.NewTicker(sseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			if err != nil {
				c.Logger().Error("SSE error: ", err)
			}
			return nil
		case t := <-telemetryChan:
			data, _ := json.Marshal(t)
			fmt.Fprintf(c.Response(), "event: telemetry\ndata: %s\n\n", string(data))
			if flusher, ok := c.Response().Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-ticker.C:
			fmt.Fprintf(c.Response(), ": keep-alive\n\n")
			if flusher, ok := c.Response().Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// WriteTelemetry handles POST /telemetry (for device data ingestion)
func (h *TelemetryHandler) WriteTelemetry(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		DeviceSN        string  `json:"device_sn" validate:"required"`
		DeviceType      string  `json:"device_type"`
		LocationName    string  `json:"location_name"`
		Temperature     float64 `json:"temperature"`
		Humidity        float64 `json:"humidity"`
		AirQualityIndex int     `json:"aqi"`
		PM25            float64 `json:"pm25"`
		PM10            float64 `json:"pm10"`
		CO2             int     `json:"co2"`
		VOC             float64 `json:"voc"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	t := &telemetry.Telemetry{
		DeviceSN:        req.DeviceSN,
		DeviceType:      req.DeviceType,
		LocationName:    req.LocationName,
		Temperature:     req.Temperature,
		Humidity:        req.Humidity,
		AirQualityIndex: req.AirQualityIndex,
		PM25:            req.PM25,
		PM10:            req.PM10,
		CO2:             req.CO2,
		VOC:             req.VOC,
		Timestamp:       time.Now(),
	}

	if err := h.svc.WriteTelemetry(ctx, t); err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, nil, "Telemetry data received successfully")
}
