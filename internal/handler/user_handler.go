package handler

import (
	"net/http"

	"aether-node/internal/domain/user"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	svc user.UserService
}

func NewUserHandler(svc user.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// CreateUser handles POST /user
func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=8"`
		FirstName string `json:"first_name" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	createReq := &user.CreateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	u, err := h.svc.CreateUser(ctx, createReq)
	if err != nil {
		if err == user.ErrEmailAlreadyExists {
			return response.Error(c, http.StatusConflict, "DUPLICATE_ENTRY", "Email already exists")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, u, "User created successfully")
}

// GetUser handles GET /user/:guid
func (h *UserHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	u, err := h.svc.GetUser(ctx, guid)
	if err != nil {
		if err == user.ErrUserNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "User not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, u, "User retrieved successfully")
}

// ListUsers handles POST /user/list
func (h *UserHandler) ListUsers(c echo.Context) error {
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

	params := &user.ListParams{
		Limit:  req.Limit,
		Page:   req.Page,
		Order:  req.Order,
		Sort:   req.Sort,
		Search: req.Search,
	}

	result, err := h.svc.ListUsers(ctx, params)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.SuccessWithPagination(c, http.StatusOK, result.Users, &response.Pagination{
		Page:       result.Page,
		Limit:      result.Limit,
		Total:      result.Total,
		TotalPages: result.TotalPage,
	})
}

// UpdateUser handles PATCH /user/:guid
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	var req struct {
		Email     *string `json:"email"`
		Password  *string `json:"password"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	updateReq := &user.UpdateUserRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	u, err := h.svc.UpdateUser(ctx, guid, updateReq)
	if err != nil {
		if err == user.ErrUserNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "User not found")
		}
		if err == user.ErrEmailAlreadyExists {
			return response.Error(c, http.StatusConflict, "DUPLICATE_ENTRY", "Email already exists")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, u, "User updated successfully")
}

// DeleteUser handles DELETE /user/:guid
func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	guid := c.Param("guid")

	err := h.svc.DeleteUser(ctx, guid)
	if err != nil {
		if err == user.ErrUserNotFound {
			return response.Error(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "User not found")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "User deleted successfully")
}
