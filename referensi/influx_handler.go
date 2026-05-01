// handlers/influx_handler.go
package handlers

import (
	"encoding/json"
	"net/http"

	"be-go-historian/internal/domain"
	"be-go-historian/internal/middleware"
	"be-go-historian/internal/services"

	"github.com/labstack/echo/v4"
)

type InfluxHandler struct {
    influxService  services.InfluxService
    deviceService  services.DeviceService
    projectService services.ProjectService
}

func NewInfluxHandler(influxService services.InfluxService, deviceService services.DeviceService, projectService services.ProjectService) *InfluxHandler {
    return &InfluxHandler{
        influxService:  influxService,
        deviceService:  deviceService,
        projectService: projectService,
    }
}

// StreamHealth streams all health data via SSE
// @Summary Stream global health data
// @Description Stream health information for all devices in the system via SSE
// @Tags monitoring
// @Produce text/event-stream
// @Security JWTAuth
// @Success 200 {string} string "SSE Stream"
// @Router /health/stream [get]
func (h *InfluxHandler) StreamHealth(c echo.Context) error {
    // Extract tenant context from JWT
    tenantID, err := middleware.GetTenantIDFromContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authentication")
    }
    isSuperAdmin := middleware.IsSuperAdmin(c)

    // Header SSE
    c.Response().Header().Set("Content-Type", "text/event-stream")
    c.Response().Header().Set("Cache-Control", "no-cache")
    c.Response().Header().Set("Connection", "keep-alive")
    c.Response().Header().Set("Access-Control-Allow-Origin", "*")

    ctx := c.Request().Context()
    dataCh := make(chan []domain.HealthData)

    go h.influxService.StreamHealth(ctx, tenantID, isSuperAdmin, dataCh)

    enc := json.NewEncoder(c.Response())

    for {
        select {
        case <-ctx.Done():
            return nil
        case data := <-dataCh:
            if _, err := c.Response().Write([]byte("data: ")); err != nil {
                return err
            }
            if err := enc.Encode(data); err != nil {
                return err
            }
            if _, err := c.Response().Write([]byte("\n\n")); err != nil {
                return err
            }
            c.Response().Flush()
        }
    }
}

// StreamHealthByProjectName streams health data for a specific project via SSE
// @Summary Stream health by project
// @Description Stream health information filtered by project name via SSE
// @Tags monitoring
// @Produce text/event-stream
// @Security JWTAuth
// @Param projectName path string true "Project Name"
// @Success 200 {string} string "SSE Stream"
// @Router /health/project/{projectName}/stream [get]
func (h *InfluxHandler) StreamHealthByProjectName(c echo.Context) error {
    // Extract tenant context from JWT
    tenantID, err := middleware.GetTenantIDFromContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authentication")
    }
    isSuperAdmin := middleware.IsSuperAdmin(c)
    projectName := c.Param("projectName")

    // Validate project exists and user has access (authorization check)
    _, err = h.projectService.GetByNameAndTenant(c.Request().Context(), projectName, tenantID, isSuperAdmin)
    if err != nil {
        return echo.NewHTTPError(http.StatusNotFound, "Project not found")
    }

    // ====== KIRIM HEADER DULU ======
    c.Response().Header().Set("Content-Type", "text/event-stream")
    c.Response().Header().Set("Cache-Control", "no-cache")
    c.Response().Header().Set("Connection", "keep-alive")
    c.Response().Header().Set("Access-Control-Allow-Origin", "*")
    c.Response().Flush() // <--- PENTING

    ctx := c.Request().Context()
    dataCh := make(chan []domain.HealthData)

    // Query by project name (as stored in MQTT/InfluxDB)
    go h.influxService.StreamHealthByProjectName(ctx, projectName, tenantID, isSuperAdmin, dataCh)

    enc := json.NewEncoder(c.Response())

    // ===== SSE LOOP =====
    for {
        select {
        case <-ctx.Done():
            return nil

        case data := <-dataCh:
            // Kirim heartbeat dulu
            c.Response().Write([]byte(": ping\n\n"))
            c.Response().Flush()

            // Baru kirim data
            c.Response().Write([]byte("data: "))
            enc.Encode(data)
            c.Response().Write([]byte("\n\n"))
            c.Response().Flush()
        }
    }
}


// StreamHealthAllProject streams a grouped health summary for all projects via SSE
// @Summary Stream health summary for all projects
// @Description Stream real-time health status summary grouped by project via SSE
// @Tags monitoring
// @Produce text/event-stream
// @Security JWTAuth
// @Success 200 {string} string "SSE Stream"
// @Router /health/grouped-project/stream [get]
func (h *InfluxHandler) StreamHealthAllProject(c echo.Context) error {
    // Extract tenant context from JWT
    tenantID, err := middleware.GetTenantIDFromContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authentication")
    }
    isSuperAdmin := middleware.IsSuperAdmin(c)

    // Headers Wajib
    c.Response().Header().Set("Content-Type", "text/event-stream")
    c.Response().Header().Set("Cache-Control", "no-cache")
    c.Response().Header().Set("Connection", "keep-alive")
    c.Response().Header().Set("Access-Control-Allow-Origin", "*")
    c.Response().Flush() // <--- PENTING

    // Pastikan flusher tersedia
    res := c.Response()
    flusher, ok := c.Response().Writer.(http.Flusher)
    if !ok {
        return echo.NewHTTPError(500, "Streaming unsupported")
    }

    ctx := c.Request().Context()
    dataCh := make(chan []domain.HealthData)

    go h.influxService.StreamHealthAllProject(ctx, tenantID, isSuperAdmin, dataCh)

    for {
        select {
        case <-ctx.Done():
            return nil
        case data := <-dataCh:
            // Encode JSON manual (lebih aman untuk SSE)
            b, err := json.Marshal(data)
            if err != nil {
                return err
            }

            // Kirim event SSE
            _, err = res.Write([]byte("data: " + string(b) + "\n\n"))
            if err != nil {
                return err
            }

            // FLUSH
            flusher.Flush()
        }
    }
}

// StreamTelemetryBySN streams real-time telemetry data for a specific device serial number via SSE
// @Summary Stream telemetry by SN
// @Description Stream sensor telemetry information for a specific device via SSE
// @Tags monitoring
// @Produce text/event-stream
// @Security JWTAuth
// @Security APIKeyAuth
// @Param sn path string true "Serial Number"
// @Success 200 {string} string "SSE Stream"
// @Router /telemetry/sn/{sn} [get]
func (h *InfluxHandler) StreamTelemetryBySN(c echo.Context) error {
    // Extract tenant context from JWT
    tenantID, err := middleware.GetTenantIDFromContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authentication")
    }
    isSuperAdmin := middleware.IsSuperAdmin(c)
    sn := c.Param("sn")

    // Validate device ownership: lookup device by SN to verify tenant access
    // This prevents users from accessing other tenants' device data
    ctx := c.Request().Context()
    device, err := h.deviceService.GetDeviceBySN(ctx, sn, tenantID, isSuperAdmin)
    if err != nil {
        return echo.NewHTTPError(http.StatusNotFound, "Device not found or access denied")
    }

    // Use the device's tenant_id for InfluxDB filtering (not requester's tenantID)
    // This ensures consistency when super admin accesses specific device
    deviceTenantID := device.TenantID

    // Set SSE headers
    res := c.Response()
    res.Header().Set("Content-Type", "text/event-stream")
    res.Header().Set("Cache-Control", "no-cache")
    res.Header().Set("Connection", "keep-alive")
    res.Header().Set("Access-Control-Allow-Origin", "*")

    // pastikan flusher tersedia
    flusher, ok := res.Writer.(http.Flusher)
    if !ok {
        return echo.NewHTTPError(http.StatusInternalServerError, "Streaming not supported")
    }

    flusher.Flush()

    dataCh := make(chan *domain.TelemetrySSE)

    // jalankan service with device's tenant context
    go h.influxService.StreamTelemetryBySN(ctx, sn, deviceTenantID, isSuperAdmin, dataCh)

    enc := json.NewEncoder(res)

    for {
        select {
        case <-ctx.Done():
            return nil

        case data, ok := <-dataCh:
            // channel closed → exit properly
            if !ok {
                return nil
            }
            // skip jika data nil (jangan kirim empty event)
            if data == nil {
                continue
            }

            // tulis data
            if _, err := res.Write([]byte("data: ")); err != nil {
                return err
            }

            if err := enc.Encode(data); err != nil {
                return err
            }

            if _, err := res.Write([]byte("\n\n")); err != nil {
                return err
            }

            flusher.Flush()
        }
    }
}

// GetQueryHistoryTelemetryBySN retrieves historical telemetry data
// @Summary Get historical telemetry
// @Description Fetch history sensor data for a specific device serial number
// @Tags monitoring
// @Produce json
// @Security JWTAuth
// @Security APIKeyAuth
// @Param sn path string true "Serial Number"
// @Param start query string false "Start time (e.g. -24h, -7d or RFC3339)"
// @Param stop query string false "Stop time (e.g. now(), -1h)"
// @Success 200 {object} []domain.TelemetryHistory
// @Failure 400 {object} domain.APIResponse
// @Router /telemetry/history/sn/{sn} [get]
func (h *InfluxHandler) GetQueryHistoryTelemetryBySN(c echo.Context) error {
    // Extract tenant context from JWT
    tenantID, err := middleware.GetTenantIDFromContext(c)
    if err != nil {
        return echo.NewHTTPError(http.StatusUnauthorized, "Invalid authentication")
    }
    isSuperAdmin := middleware.IsSuperAdmin(c)
    sn := c.Param("sn")

    // Validate device ownership
    ctx := c.Request().Context()
    device, err := h.deviceService.GetDeviceBySN(ctx, sn, tenantID, isSuperAdmin)
    if err != nil {
        return echo.NewHTTPError(http.StatusNotFound, "Device not found or access denied")
    }

    // Use device's tenant_id for filtering
    deviceTenantID := device.TenantID

    start := c.QueryParam("start")
    stop := c.QueryParam("stop")

    data, err := h.influxService.GetQueryHistoryTelemetryBySN(ctx, sn, start, stop, deviceTenantID, isSuperAdmin)
    if err != nil {
        return err
    }
    return c.JSON(http.StatusOK, data)
}