package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// HealthChecker pings external dependencies
type HealthChecker struct {
	pool      *pgxpool.Pool
	influxURL string
	influxToken string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(pool *pgxpool.Pool, influxURL, influxToken string) *HealthChecker {
	return &HealthChecker{
		pool:        pool,
		influxURL:   influxURL,
		influxToken: influxToken,
	}
}

// ComponentStatus represents the health status of a single component
type ComponentStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse represents the full health check response
type HealthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]ComponentStatus `json:"components"`
}

// Check performs health checks on all components
func (h *HealthChecker) Check(ctx context.Context) *HealthResponse {
	resp := &HealthResponse{
		Status:     "healthy",
		Components: make(map[string]ComponentStatus),
	}

	// Check PostgreSQL
	postgresStatus := h.checkPostgres(ctx)
	resp.Components["postgres"] = postgresStatus
	if postgresStatus.Status != "up" {
		resp.Status = "degraded"
	}

	// Check InfluxDB (non-blocking, best effort)
	influxStatus := h.checkInfluxDB(ctx)
	resp.Components["influxdb"] = influxStatus
	if influxStatus.Status != "up" && influxStatus.Error != "" {
		// InfluxDB failure doesn't make the whole system unhealthy
		// since it might be temporarily unavailable
		if resp.Status == "healthy" {
			resp.Status = "degraded"
		}
	}

	return resp
}

// checkPostgres pings PostgreSQL and returns component status
func (h *HealthChecker) checkPostgres(ctx context.Context) ComponentStatus {
	start := time.Now()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := h.pool.Ping(pingCtx)
	latency := time.Since(start)

	if err != nil {
		return ComponentStatus{
			Status:  "down",
			Latency: latency.String(),
			Error:   err.Error(),
		}
	}

	return ComponentStatus{
		Status:  "up",
		Latency: latency.String(),
	}
}

// checkInfluxDB pings InfluxDB via HTTP API
func (h *HealthChecker) checkInfluxDB(ctx context.Context) ComponentStatus {
	if h.influxURL == "" {
		return ComponentStatus{
			Status: "disabled",
		}
	}

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.influxURL+"/health", nil)
	if err != nil {
		return ComponentStatus{
			Status:  "down",
			Latency: time.Since(start).String(),
			Error:   err.Error(),
		}
	}

	if h.influxToken != "" {
		req.Header.Set("Authorization", "Token "+h.influxToken)
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return ComponentStatus{
			Status:  "down",
			Latency: latency.String(),
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ComponentStatus{
			Status:  "down",
			Latency: latency.String(),
			Error:   "HTTP " + resp.Status,
		}
	}

	return ComponentStatus{
		Status:  "up",
		Latency: latency.String(),
	}
}

// HealthHandler handles health check requests
type HealthHandler struct {
	checker *HealthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(checker *HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// GetHealth handles GET /health
func (h *HealthHandler) GetHealth(c echo.Context) error {
	ctx := c.Request().Context()
	health := h.checker.Check(ctx)

	// Determine HTTP status code
	statusCode := http.StatusOK
	if health.Status == "down" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, health)
}

// Liveness handles GET /health/live - basic liveness probe
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "alive"})
}

// Readiness handles GET /health/ready - readiness probe (checks dependencies)
func (h *HealthHandler) Readiness(c echo.Context) error {
	ctx := c.Request().Context()
	health := h.checker.Check(ctx)

	statusCode := http.StatusOK
	if health.Status == "down" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, health)
}
