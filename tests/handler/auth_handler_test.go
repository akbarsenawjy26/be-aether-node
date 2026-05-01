package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aether-node/internal/domain/auth"
	"aether-node/internal/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of auth.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.LoginResult), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, userGUID string) error {
	args := m.Called(ctx, userGUID)
	return args.Error(0)
}

func (m *MockAuthService) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.UserInfo, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserInfo), args.Error(1)
}

func (m *MockAuthService) ForgotPassword(ctx context.Context, req *auth.ForgotPasswordRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.LoginResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.LoginResult), args.Error(1)
}

func (m *MockAuthService) ValidateAccessToken(ctx context.Context, tokenString string) (string, error) {
	args := m.Called(ctx, tokenString)
	return args.String(0), args.Error(1)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	reqBody := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedResult := &auth.LoginResult{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresIn:    3600,
	}

	mockSvc.On("Login", mock.Anything, mock.MatchedBy(func(req *auth.LoginRequest) bool {
		return req.Email == "test@example.com" && req.Password == "password123"
	})).Return(expectedResult, nil)

	// Execute
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Login successful", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	reqBody := `{"email":"test@example.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("Login", mock.Anything, mock.Anything).Return(nil, auth.ErrInvalidCredentials)

	// Execute
	err := h.Login(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "AUTH_INVALID_CREDENTIALS", response["error"].(map[string]interface{})["code"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Register_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	reqBody := `{"email":"newuser@example.com","password":"password123","first_name":"New","last_name":"User"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedUserInfo := &auth.UserInfo{
		GUID:      "new-user-guid-123",
		Email:     "newuser@example.com",
		FirstName: "New",
		LastName:  "User",
	}

	mockSvc.On("Register", mock.Anything, mock.MatchedBy(func(req *auth.RegisterRequest) bool {
		return req.Email == "newuser@example.com" &&
			req.FirstName == "New" &&
			req.LastName == "User"
	})).Return(expectedUserInfo, nil)

	// Execute
	err := h.Register(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Registration successful", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	reqBody := `{"refresh_token":"old-refresh-token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/token/refresh", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedResult := &auth.LoginResult{
		AccessToken:  "new-access-token-123",
		RefreshToken: "new-refresh-token-456",
		ExpiresIn:    3600,
	}

	mockSvc.On("RefreshToken", mock.Anything, mock.MatchedBy(func(req *auth.RefreshTokenRequest) bool {
		return req.RefreshToken == "old-refresh-token"
	})).Return(expectedResult, nil)

	// Execute
	err := h.RefreshToken(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Token refreshed successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidToken(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	reqBody := `{"refresh_token":"invalid-token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/token/refresh", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("RefreshToken", mock.Anything, mock.Anything).Return(nil, auth.ErrUserNotFound)

	// Execute
	err := h.RefreshToken(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "AUTH_REFRESH_TOKEN_INVALID", response["error"].(map[string]interface{})["code"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_guid", "user-guid-123")

	mockSvc.On("Logout", mock.Anything, "user-guid-123").Return(nil)

	// Execute
	err := h.Logout(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Logged out successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestAuthHandler_Logout_InvalidToken(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Don't set user_guid - simulating invalid token

	// Execute
	err := h.Logout(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "AUTH_TOKEN_INVALID", response["error"].(map[string]interface{})["code"])
}

func TestAuthHandler_ForgotPassword_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockAuthService)
	h := handler.NewAuthHandler(mockSvc)

	reqBody := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// For forgot password, we always return success to prevent email enumeration
	mockSvc.On("ForgotPassword", mock.Anything, mock.Anything).Return(nil)

	// Execute
	err := h.ForgotPassword(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "Password reset instructions sent to email", response["message"])

	mockSvc.AssertExpectations(t)
}
