package handler

import (
	"net/http"

	"aether-node/internal/domain/threshold"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type ThresholdHandler struct {
	svc threshold.ThresholdService
}

func NewThresholdHandler(svc threshold.ThresholdService) *ThresholdHandler {
	return &ThresholdHandler{svc: svc}
}

func (h *ThresholdHandler) CreateThreshold(c echo.Context) error {
	ctx := c.Request().Context()
	var req threshold.CreateThresholdRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	t, err := h.svc.CreateThreshold(ctx, &req)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, t, "Threshold created successfully")
}

func (h *ThresholdHandler) GetThreshold(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	t, err := h.svc.GetThreshold(ctx, guid)
	if err != nil {
		if err == threshold.ErrThresholdNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Threshold not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, t, "Threshold retrieved successfully")
}

func (h *ThresholdHandler) ListThresholdsByDevice(c echo.Context) error {
	ctx := c.Request().Context()
	deviceGUID := c.Param("device_guid")

	results, err := h.svc.ListThresholdsByDevice(ctx, deviceGUID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, results, "Thresholds retrieved successfully")
}

func (h *ThresholdHandler) UpdateThreshold(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	var req threshold.UpdateThresholdRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	t, err := h.svc.UpdateThreshold(ctx, guid, &req)
	if err != nil {
		if err == threshold.ErrThresholdNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Threshold not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, t, "Threshold updated successfully")
}

func (h *ThresholdHandler) DeleteThreshold(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	err := h.svc.DeleteThreshold(ctx, guid)
	if err != nil {
		if err == threshold.ErrThresholdNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Threshold not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "Threshold deleted successfully")
}
