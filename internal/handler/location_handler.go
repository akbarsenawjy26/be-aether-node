package handler

import (
	"net/http"

	"aether-node/internal/domain/location"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type LocationHandler struct {
	svc location.LocationService
}

func NewLocationHandler(svc location.LocationService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

// CreateLocation handles POST /location
func (h *LocationHandler) CreateLocation(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Name  string `json:"name" validate:"required"`
		Notes string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	createReq := &location.CreateLocationRequest{
		Name:  req.Name,
		Notes: req.Notes,
	}

	loc, err := h.svc.CreateLocation(ctx, createReq)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, loc, "Location created successfully")
}

// GetLocation handles GET /location/:guid
func (h *LocationHandler) GetLocation(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	loc, err := h.svc.GetLocation(ctx, guid)
	if err != nil {
		if err == location.ErrLocationNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Location not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, loc, "Location retrieved successfully")
}

// ListLocations handles POST /location/list
func (h *LocationHandler) ListLocations(c echo.Context) error {
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

	params := &location.ListParams{
		Limit:  req.Limit,
		Page:   req.Page,
		Order:  req.Order,
		Sort:   req.Sort,
		Search: req.Search,
	}

	result, err := h.svc.ListLocations(ctx, params)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.Locations, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
}

// UpdateLocation handles PATCH /location/:guid
func (h *LocationHandler) UpdateLocation(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	var req struct {
		Name  *string `json:"name"`
		Notes *string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	updateReq := &location.UpdateLocationRequest{
		Name:  req.Name,
		Notes: req.Notes,
	}

	loc, err := h.svc.UpdateLocation(ctx, guid, updateReq)
	if err != nil {
		if err == location.ErrLocationNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Location not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, loc, "Location updated successfully")
}

// DeleteLocation handles DELETE /location/:guid
func (h *LocationHandler) DeleteLocation(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	err := h.svc.DeleteLocation(ctx, guid)
	if err != nil {
		if err == location.ErrLocationNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "Location not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "Location deleted successfully")
}
