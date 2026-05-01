package handler

import (
	"net/http"

	"aether-node/internal/domain/device"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type DeviceHandler struct {
	svc device.DeviceService
}

func NewDeviceHandler(svc device.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// CreateDevice handles POST /device
func (h *DeviceHandler) CreateDevice(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Type         string `json:"type" validate:"required"`
		SerialNumber string `json:"serial_number" validate:"required"`
		Alias        string `json:"alias"`
		Notes        string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	createReq := &device.CreateDeviceRequest{
		Type:         req.Type,
		SerialNumber: req.SerialNumber,
		Alias:        req.Alias,
		Notes:        req.Notes,
	}

	d, err := h.svc.CreateDevice(ctx, createReq)
	if err != nil {
		if err == device.ErrDeviceSerialNumberExists {
			return response.Error(c, http.StatusConflict, "DUPLICATE_ENTRY", "Serial number already exists")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, d, "Device created successfully")
}

// GetDevice handles GET /device/:guid
func (h *DeviceHandler) GetDevice(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	d, err := h.svc.GetDevice(ctx, guid)
	if err != nil {
		if err == device.ErrDeviceNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Device not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, d, "Device retrieved successfully")
}

// ListDevices handles POST /device/list
func (h *DeviceHandler) ListDevices(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Limit  int    `json:"limit"`
		Page   int    `json:"page"`
		Order  string `json:"order"`
		Sort   string `json:"sort"`
		Search string `json:"search"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	params := &device.ListParams{
		Limit:  req.Limit,
		Page:   req.Page,
		Order:  req.Order,
		Sort:   req.Sort,
		Search: req.Search,
	}

	result, err := h.svc.ListDevices(ctx, params)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.Devices, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
}

// UpdateDevice handles PATCH /device/:guid
func (h *DeviceHandler) UpdateDevice(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	var req struct {
		Type         *string `json:"type"`
		SerialNumber *string `json:"serial_number"`
		Alias        *string `json:"alias"`
		Notes        *string `json:"notes"`
		IsActive     *bool   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	updateReq := &device.UpdateDeviceRequest{
		Type:         req.Type,
		SerialNumber: req.SerialNumber,
		Alias:        req.Alias,
		Notes:        req.Notes,
		IsActive:     req.IsActive,
	}

	d, err := h.svc.UpdateDevice(ctx, guid, updateReq)
	if err != nil {
		if err == device.ErrDeviceNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Device not found")
		}
		if err == device.ErrDeviceSerialNumberExists {
			return response.Error(c, http.StatusConflict, "DUPLICATE_ENTRY", "Serial number already exists")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, d, "Device updated successfully")
}

// DeleteDevice handles DELETE /device/:guid
func (h *DeviceHandler) DeleteDevice(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	err := h.svc.DeleteDevice(ctx, guid)
	if err != nil {
		if err == device.ErrDeviceNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Device not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "Device deleted successfully")
}
