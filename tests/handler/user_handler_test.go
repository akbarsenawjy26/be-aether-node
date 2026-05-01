package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aether-node/internal/domain/user"
	"aether-node/internal/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of user.UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, guid string) (*user.User, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, params *user.ListParams) (*user.ListResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.ListResult), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, guid string, req *user.UpdateUserRequest) (*user.User, error) {
	args := m.Called(ctx, guid, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, guid string) error {
	args := m.Called(ctx, guid)
	return args.Error(0)
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	// Request body
	reqBody := `{"email":"test@example.com","password":"password123","first_name":"John","last_name":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Expected user
	expectedUser := &user.User{
		GUID:         "test-guid-123",
		Email:        "test@example.com",
		FirstName:    "John",
		LastName:     "Doe",
		IsActive:     true,
		PasswordHash: "$2a$10...", // hashed password
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Mock expectations
	mockSvc.On("CreateUser", mock.Anything, mock.MatchedBy(func(req *user.CreateUserRequest) bool {
		return req.Email == "test@example.com" &&
			req.FirstName == "John" &&
			req.LastName == "Doe"
	})).Return(expectedUser, nil)

	// Execute
	err := h.CreateUser(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "User created successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_CreateUser_ValidationError(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	// Request body with invalid email
	reqBody := `{"email":"invalid-email","password":"password123","first_name":"John","last_name":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock expectations (service should still be called, but validation happens at service level)
	mockSvc.On("CreateUser", mock.Anything, mock.Anything).Return(nil, user.ErrEmailAlreadyExists)

	// Execute
	err := h.CreateUser(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "DUPLICATE_ENTRY", response["error"].(map[string]interface{})["code"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/user/test-guid-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("guid")
	c.SetParamValues("test-guid-123")

	expectedUser := &user.User{
		GUID:      "test-guid-123",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		IsActive:  true,
	}

	mockSvc.On("GetUser", mock.Anything, "test-guid-123").Return(expectedUser, nil)

	// Execute
	err := h.GetUser(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "User retrieved successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/user/nonexistent-guid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("guid")
	c.SetParamValues("nonexistent-guid")

	mockSvc.On("GetUser", mock.Anything, "nonexistent-guid").Return(nil, user.ErrUserNotFound)

	// Execute
	err := h.GetUser(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, false, response["success"])
	assert.Equal(t, "RESOURCE_NOT_FOUND", response["error"].(map[string]interface{})["code"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_ListUsers_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	reqBody := `{"limit":10,"page":1,"search":"john"}`
	req := httptest.NewRequest(http.MethodPost, "/user/list", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedUsers := []*user.User{
		{GUID: "guid-1", Email: "john@example.com", FirstName: "John", LastName: "Doe"},
		{GUID: "guid-2", Email: "johnny@example.com", FirstName: "Johnny", LastName: "Smith"},
	}

	expectedResult := &user.ListResult{
		Users:     expectedUsers,
		Total:     2,
		Page:      1,
		Limit:     10,
		TotalPage: 1,
	}

	mockSvc.On("ListUsers", mock.Anything, mock.MatchedBy(func(params *user.ListParams) bool {
		return params.Limit == 10 && params.Page == 1 && params.Search == "john"
	})).Return(expectedResult, nil)

	// Execute
	err := h.ListUsers(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/user/test-guid-123", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("guid")
	c.SetParamValues("test-guid-123")

	mockSvc.On("DeleteUser", mock.Anything, "test-guid-123").Return(nil)

	// Execute
	err := h.DeleteUser(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "User deleted successfully", response["message"])

	mockSvc.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_NotFound(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockUserService)
	h := handler.NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodDelete, "/user/nonexistent-guid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("guid")
	c.SetParamValues("nonexistent-guid")

	mockSvc.On("DeleteUser", mock.Anything, "nonexistent-guid").Return(user.ErrUserNotFound)

	// Execute
	err := h.DeleteUser(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockSvc.AssertExpectations(t)
}
