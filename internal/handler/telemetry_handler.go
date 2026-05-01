package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"aether-node/internal/domain/telemetry"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type TelemetryHandler struct {
	svc telemetry.TelemetryService
}

func NewTelemetryHandler(svc telemetry.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{svc: svc}
}

// StreamAllDevices handles GET /stream - SSE for all devices
func (h *TelemetryHandler) StreamAllDevices(c echo.Context) error {
	ctx := c.Request().Context()

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().WriteHeader(http.StatusOK)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	telemetryChan, errChan := h.svc.StreamAllDevices(ctx)
	ticker := time.NewTicker(30 * time.Second)
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

// StreamDevice handles GET /stream/:device-sn - SSE for specific device
func (h *TelemetryHandler) StreamDevice(c echo.Context) error {
	ctx := c.Request().Context()
	deviceSN := c.Param("device-sn")

	if deviceSN == "" {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Device serial number is required")
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().WriteHeader(http.StatusOK)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	telemetryChan, errChan := h.svc.StreamDevice(ctx, deviceSN)
	ticker := time.NewTicker(30 * time.Second)
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

// GetHistory handles POST /history/telemetry/:device-sn
func (h *TelemetryHandler) GetHistory(c echo.Context) error {
	ctx := c.Request().Context()
	deviceSN := c.Param("device-sn")

	if deviceSN == "" {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Device serial number is required")
	}

	var req struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Limit     int    `json:"limit"`
		Page      int    `json:"page"`
		Order     string `json:"order"`
		Sort      string `json:"sort"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	query := &telemetry.TelemetryQuery{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Limit:     req.Limit,
		Page:      req.Page,
		Order:     req.Order,
		Sort:      req.Sort,
	}

	result, err := h.svc.GetHistory(ctx, deviceSN, query)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.Telemetry, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
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
