package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aether-node/internal/domain/telemetry"
	"aether-node/internal/handler"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTelemetryService is a mock implementation of telemetry.TelemetryService
type MockTelemetryService struct {
	mock.Mock
}

func (m *MockTelemetryService) WriteTelemetry(ctx context.Context, t *telemetry.Telemetry) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTelemetryService) StreamAllDevices(ctx context.Context) (<-chan *telemetry.Telemetry, <-chan error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan *telemetry.Telemetry), args.Get(1).(<-chan error)
}

func (m *MockTelemetryService) StreamDevice(ctx context.Context, deviceSN string) (<-chan *telemetry.Telemetry, <-chan error) {
	args := m.Called(ctx, deviceSN)
	return args.Get(0).(<-chan *telemetry.Telemetry), args.Get(1).(<-chan error)
}

func (m *MockTelemetryService) GetHistory(ctx context.Context, deviceSN string, query *telemetry.TelemetryQuery) (*telemetry.TelemetryListResult, error) {
	args := m.Called(ctx, deviceSN, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telemetry.TelemetryListResult), args.Error(1)
}

func (m *MockTelemetryService) StreamAllDevicesWithHealth(ctx context.Context, project string) (telemetry.DevicePayload, error) {
	args := m.Called(ctx, project)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(telemetry.DevicePayload), args.Error(1)
}

func (m *MockTelemetryService) StreamDeviceWithHealth(ctx context.Context, project, deviceSN string) (*telemetry.DeviceEntry, error) {
	args := m.Called(ctx, project, deviceSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telemetry.DeviceEntry), args.Error(1)
}

func (m *MockTelemetryService) GetTelemetryHistory(ctx context.Context, filter telemetry.HistoryFilter) ([]telemetry.TelemetryRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telemetry.TelemetryRecord), args.Error(1)
}

// ============================================================
// Handler Tests
// ============================================================

func TestTelemetryHandler_GetTelemetryHistory_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	reqBody := `{"start":"2024-01-01T00:00:00Z","stop":"2024-01-31T23:59:59Z","window":"5m"}`
	req := httptest.NewRequest(http.MethodPost, "/telemetry/history/SN001", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/telemetry/history/:device_sn")
	c.SetParamNames("device_sn")
	c.SetParamValues("SN001")
	c.QueryParams().Set("project", "project-a")

	expectedRecords := []telemetry.TelemetryRecord{
		{
			Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Fields:    map[string]interface{}{"temperature": 25.5, "humidity": 60.0},
		},
		{
			Timestamp: time.Date(2024, 1, 1, 0, 5, 0, 0, time.UTC),
			Fields:    map[string]interface{}{"temperature": 26.0, "humidity": 61.0},
		},
	}

	mockSvc.On("GetTelemetryHistory", mock.Anything, mock.MatchedBy(func(filter telemetry.HistoryFilter) bool {
		return filter.DeviceSN == "SN001" &&
			filter.Project == "project-a" &&
			filter.Window == "5m"
	})).Return(expectedRecords, nil)

	// Execute
	err := h.GetTelemetryHistory(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "SN001", response["device_sn"])

	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	mockSvc.AssertExpectations(t)
}

func TestTelemetryHandler_GetTelemetryHistory_MissingStart(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	reqBody := `{"stop":"2024-01-31T23:59:59Z"}`
	req := httptest.NewRequest(http.MethodPost, "/telemetry/history/SN001", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/telemetry/history/:device_sn")
	c.SetParamNames("device_sn")
	c.SetParamValues("SN001")

	// Execute
	err := h.GetTelemetryHistory(c)

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestTelemetryHandler_GetTelemetryHistory_InvalidStartFormat(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	reqBody := `{"start":"invalid-date","stop":"2024-01-31T23:59:59Z"}`
	req := httptest.NewRequest(http.MethodPost, "/telemetry/history/SN001", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/telemetry/history/:device_sn")
	c.SetParamNames("device_sn")
	c.SetParamValues("SN001")

	// Execute
	err := h.GetTelemetryHistory(c)

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestTelemetryHandler_GetTelemetryHistory_StopBeforeStart(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	reqBody := `{"start":"2024-01-31T00:00:00Z","stop":"2024-01-01T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/telemetry/history/SN001", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/telemetry/history/:device_sn")
	c.SetParamNames("device_sn")
	c.SetParamValues("SN001")

	// Execute
	err := h.GetTelemetryHistory(c)

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

func TestTelemetryHandler_StreamDeviceWithHealth_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/stream/SN001?project=project-a", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/telemetry/stream/:device_sn")
	c.SetParamNames("device_sn")
	c.SetParamValues("SN001")

	expectedEntry := &telemetry.DeviceEntry{
		Health: telemetry.HealthData{
			DeviceSN:  "SN001",
			Project:   "project-a",
			Type:      "basic-model",
			Status:    "online",
			Uptime:    3600,
			Temp:      25.5,
			Hum:       60.0,
			LastSeen:  time.Now(),
		},
		Telemetry: telemetry.TelemetryData{
			"temperature": 25.5,
			"humidity":    60.0,
		},
	}

	mockSvc.On("StreamDeviceWithHealth", mock.Anything, "project-a", "SN001").Return(expectedEntry, nil)

	// Execute - SSE endpoint sends data and sets headers
	_ = h.StreamDevice(c)

	// Assert - SSE doesn't return error, it flushes data
	// Headers should be set for SSE
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestTelemetryHandler_StreamDeviceWithHealth_DeviceNotFound(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/stream/UNKNOWN", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/telemetry/stream/:device_sn")
	c.SetParamNames("device_sn")
	c.SetParamValues("UNKNOWN")

	mockSvc.On("StreamDeviceWithHealth", mock.Anything, "", "UNKNOWN").Return(nil, nil)

	// Execute
	_ = h.StreamDevice(c)

	// Assert - SSE endpoint
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	// Note: The handler sends error via SSE event, doesn't return HTTP error
}

func TestTelemetryHandler_StreamAllDevicesWithHealth_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/stream?project=project-a", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedPayload := telemetry.DevicePayload{
		"basic-model": []telemetry.DeviceEntry{
			{
				Health: telemetry.HealthData{
					DeviceSN: "SN001",
					Type:     "basic-model",
					Status:   "online",
				},
				Telemetry: telemetry.TelemetryData{
					"temperature": 25.5,
				},
			},
		},
		"aqi-model": []telemetry.DeviceEntry{
			{
				Health: telemetry.HealthData{
					DeviceSN: "SN002",
					Type:     "aqi-model",
					Status:   "offline",
				},
				Telemetry: telemetry.TelemetryData{},
			},
		},
	}

	mockSvc.On("StreamAllDevicesWithHealth", mock.Anything, "project-a").Return(expectedPayload, nil)

	// Execute
	_ = h.StreamAllDevices(c)

	// Assert
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rec.Code)

	mockSvc.AssertExpectations(t)
}

func TestTelemetryHandler_WriteTelemetry_Success(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	reqBody := `{
		"device_sn": "SN001",
		"device_type": "basic-model",
		"location_name": "Living Room",
		"temperature": 25.5,
		"humidity": 60.0,
		"aqi": 50,
		"pm25": 10.5,
		"pm10": 20.0,
		"co2": 400,
		"voc": 0.5
	}`
	req := httptest.NewRequest(http.MethodPost, "/telemetry", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockSvc.On("WriteTelemetry", mock.Anything, mock.MatchedBy(func(t *telemetry.Telemetry) bool {
		return t.DeviceSN == "SN001" &&
			t.DeviceType == "basic-model" &&
			t.Temperature == 25.5
	})).Return(nil)

	// Execute
	err := h.WriteTelemetry(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])

	mockSvc.AssertExpectations(t)
}

func TestTelemetryHandler_WriteTelemetry_InvalidJSON(t *testing.T) {
	// Setup
	e := echo.New()
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/telemetry", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	err := h.WriteTelemetry(c)

	// Assert
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
}

// ============================================================
// Service Tests
// ============================================================

// MockTelemetryRepository is a mock implementation of telemetry.TelemetryRepository
type MockTelemetryRepository struct {
	mock.Mock
}

func (m *MockTelemetryRepository) WriteTelemetry(ctx context.Context, t *telemetry.Telemetry) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockTelemetryRepository) GetLatestByDeviceSN(ctx context.Context, deviceSN string) (*telemetry.Telemetry, error) {
	args := m.Called(ctx, deviceSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telemetry.Telemetry), args.Error(1)
}

func (m *MockTelemetryRepository) GetAllLatest(ctx context.Context) ([]*telemetry.Telemetry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*telemetry.Telemetry), args.Error(1)
}

func (m *MockTelemetryRepository) QueryHistory(ctx context.Context, deviceSN string, query *telemetry.TelemetryQuery) (*telemetry.TelemetryListResult, error) {
	args := m.Called(ctx, deviceSN, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*telemetry.TelemetryListResult), args.Error(1)
}

func (m *MockTelemetryRepository) GetLatestHealth(ctx context.Context, filter telemetry.DeviceFilter) ([]telemetry.HealthData, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telemetry.HealthData), args.Error(1)
}

func (m *MockTelemetryRepository) GetLatestTelemetry(ctx context.Context, filter telemetry.DeviceFilter) (map[string]telemetry.TelemetryData, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]telemetry.TelemetryData), args.Error(1)
}

func (m *MockTelemetryRepository) GetTelemetryHistory(ctx context.Context, filter telemetry.HistoryFilter) ([]telemetry.TelemetryRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]telemetry.TelemetryRecord), args.Error(1)
}

// Helper to create mock service for testing
func newMockTelemetryService() (*MockTelemetryService, *handler.TelemetryHandler) {
	mockSvc := new(MockTelemetryService)
	h := handler.NewTelemetryHandler(mockSvc)
	return mockSvc, h
}
