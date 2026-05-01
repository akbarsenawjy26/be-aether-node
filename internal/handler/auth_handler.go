package handler

import (
	"net/http"

	"aether-node/internal/domain/auth"
	"aether-node/pkg/response"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	svc auth.AuthService
}

func NewAuthHandler(svc auth.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	loginReq := &auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	result, err := h.svc.Login(ctx, loginReq)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			return response.Error(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid email or password")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, result, "Login successful")
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	// Get user GUID from context (set by auth middleware)
	userGUID, ok := c.Get("user_guid").(string)
	if !ok {
		return response.Error(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	err := h.svc.Logout(ctx, userGUID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusOK, nil, "Logged out successfully")
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c echo.Context) error {
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

	registerReq := &auth.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	userInfo, err := h.svc.Register(ctx, registerReq)
	if err != nil {
		if err == auth.ErrUserNotFound {
			// This shouldn't happen for registration, but handle it
			return response.Error(c, http.StatusConflict, "DUPLICATE_ENTRY", "Email already exists")
		}
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, http.StatusCreated, userInfo, "Registration successful")
}

// ForgotPassword handles POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	forgotReq := &auth.ForgotPasswordRequest{
		Email: req.Email,
	}

	err := h.svc.ForgotPassword(ctx, forgotReq)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	// Always return success to prevent email enumeration
	return response.Success(c, http.StatusOK, nil, "Password reset instructions sent to email")
}

// RefreshToken handles POST /auth/token/refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()

	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	refreshReq := &auth.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	result, err := h.svc.RefreshToken(ctx, refreshReq)
	if err != nil {
		return response.Error(c, http.StatusUnauthorized, "AUTH_REFRESH_TOKEN_INVALID", "Invalid or expired refresh token")
	}

	return response.Success(c, http.StatusOK, result, "Token refreshed successfully")
}
