package handler

import (
	"net/http"

	"aether-node/internal/domain/apikey"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type APIKeyHandler struct {
	svc apikey.APIKeyService
}

func NewAPIKeyHandler(svc apikey.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{svc: svc}
}

// CreateAPIKey handles POST /apikey
func (h *APIKeyHandler) CreateAPIKey(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Notes      string `json:"notes"`
		ExpireDate string `json:"expire_date" validate:"required"`
		IsActive   bool   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	createReq := &apikey.CreateAPIKeyRequest{
		Notes:      req.Notes,
		ExpireDate: req.ExpireDate,
		IsActive:   req.IsActive,
	}

	apiKeyWithKey, err := h.svc.CreateAPIKey(ctx, createReq)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, apiKeyWithKey, "API key created successfully. Please save this key - it will not be shown again.")
}

// GetAPIKey handles GET /apikey/:guid
func (h *APIKeyHandler) GetAPIKey(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	apiKey, err := h.svc.GetAPIKey(ctx, guid)
	if err != nil {
		if err == apikey.ErrAPIKeyNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "API key not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, apiKey, "API key retrieved successfully")
}

// ListAPIKeys handles POST /apikey/list
func (h *APIKeyHandler) ListAPIKeys(c echo.Context) error {
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

	params := &apikey.ListParams{
		Limit:  req.Limit,
		Page:   req.Page,
		Order:  req.Order,
		Sort:   req.Sort,
		Search: req.Search,
	}

	result, err := h.svc.ListAPIKeys(ctx, params)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.APIKeys, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
}

// UpdateAPIKey handles PATCH /apikey/:guid
func (h *APIKeyHandler) UpdateAPIKey(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	var req struct {
		Notes      *string `json:"notes"`
		ExpireDate *string `json:"expire_date"`
		IsActive   *bool   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	updateReq := &apikey.UpdateAPIKeyRequest{
		Notes:      req.Notes,
		ExpireDate: req.ExpireDate,
		IsActive:   req.IsActive,
	}

	apiKey, err := h.svc.UpdateAPIKey(ctx, guid, updateReq)
	if err != nil {
		if err == apikey.ErrAPIKeyNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "API key not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, apiKey, "API key updated successfully")
}

// DeleteAPIKey handles DELETE /apikey/:guid
func (h *APIKeyHandler) DeleteAPIKey(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	err := h.svc.DeleteAPIKey(ctx, guid)
	if err != nil {
		if err == apikey.ErrAPIKeyNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "API key not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "API key deleted successfully")
}
