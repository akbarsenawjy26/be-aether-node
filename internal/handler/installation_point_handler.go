package handler

import (
	"net/http"

	"aether-node/internal/domain/installation_point"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type InstallationPointHandler struct {
	svc installation_point.InstallationPointService
}

func NewInstallationPointHandler(svc installation_point.InstallationPointService) *InstallationPointHandler {
	return &InstallationPointHandler{svc: svc}
}

// CreateInstallationPoint handles POST /installation-point
func (h *InstallationPointHandler) CreateInstallationPoint(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Name         string `json:"name" validate:"required"`
		DeviceGUID   string `json:"device_guid" validate:"required"`
		LocationGUID string `json:"location_guid" validate:"required"`
		Notes        string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	createReq := &installation_point.CreateInstallationPointRequest{
		Name:         req.Name,
		DeviceGUID:   req.DeviceGUID,
		LocationGUID: req.LocationGUID,
		Notes:        req.Notes,
	}

	ip, err := h.svc.CreateInstallationPoint(ctx, createReq)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, ip, "Installation point created successfully")
}

// GetInstallationPoint handles GET /installation-point/:guid
func (h *InstallationPointHandler) GetInstallationPoint(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	ip, err := h.svc.GetInstallationPoint(ctx, guid)
	if err != nil {
		if err == installation_point.ErrInstallationPointNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Installation point not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, ip, "Installation point retrieved successfully")
}

// GetInstallationPointWithRelations handles GET /installation-point/:guid/relations
func (h *InstallationPointHandler) GetInstallationPointWithRelations(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	ip, err := h.svc.GetInstallationPointWithRelations(ctx, guid)
	if err != nil {
		if err == installation_point.ErrInstallationPointNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Installation point not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, ip, "Installation point retrieved successfully")
}

// ListInstallationPoints handles POST /installation-point/list
func (h *InstallationPointHandler) ListInstallationPoints(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Limit         int    `json:"limit"`
		Page          int    `json:"page"`
		Order         string `json:"order"`
		Sort          string `json:"sort"`
		Search        string `json:"search"`
		DeviceGUID    string `json:"device_guid"`
		LocationGUID  string `json:"location_guid"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	params := &installation_point.ListParams{
		Limit:        req.Limit,
		Page:         req.Page,
		Order:        req.Order,
		Sort:         req.Sort,
		Search:       req.Search,
		DeviceGUID:   req.DeviceGUID,
		LocationGUID: req.LocationGUID,
	}

	result, err := h.svc.ListInstallationPoints(ctx, params)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.InstallationPoints, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
}

// UpdateInstallationPoint handles PATCH /installation-point/:guid
func (h *InstallationPointHandler) UpdateInstallationPoint(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	var req struct {
		Name         *string `json:"name"`
		DeviceGUID   *string `json:"device_guid"`
		LocationGUID *string `json:"location_guid"`
		Notes        *string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	updateReq := &installation_point.UpdateInstallationPointRequest{
		Name:         req.Name,
		DeviceGUID:   req.DeviceGUID,
		LocationGUID: req.LocationGUID,
		Notes:        req.Notes,
	}

	ip, err := h.svc.UpdateInstallationPoint(ctx, guid, updateReq)
	if err != nil {
		if err == installation_point.ErrInstallationPointNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Installation point not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, ip, "Installation point updated successfully")
}

// DeleteInstallationPoint handles DELETE /installation-point/:guid
func (h *InstallationPointHandler) DeleteInstallationPoint(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	err := h.svc.DeleteInstallationPoint(ctx, guid)
	if err != nil {
		if err == installation_point.ErrInstallationPointNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Installation point not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "Installation point deleted successfully")
}
