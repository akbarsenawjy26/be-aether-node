package service_test

import (
	"context"
	"testing"
	"time"

	domain "aether-node/internal/domain/telemetry"
	"aether-node/internal/service/telemetry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of domain.TelemetryRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) WriteTelemetry(ctx context.Context, t *domain.Telemetry) error {
	args := m.Called(ctx, t)
	return args.Error(0)
}

func (m *MockRepository) GetLatestByDeviceSN(ctx context.Context, deviceSN string) (*domain.Telemetry, error) {
	args := m.Called(ctx, deviceSN)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Telemetry), args.Error(1)
}

func (m *MockRepository) GetAllLatest(ctx context.Context) ([]*domain.Telemetry, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Telemetry), args.Error(1)
}

func (m *MockRepository) QueryHistory(ctx context.Context, deviceSN string, query *domain.TelemetryQuery) (*domain.TelemetryListResult, error) {
	args := m.Called(ctx, deviceSN, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TelemetryListResult), args.Error(1)
}

func (m *MockRepository) GetLatestHealth(ctx context.Context, filter domain.DeviceFilter) ([]domain.HealthData, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.HealthData), args.Error(1)
}

func (m *MockRepository) GetLatestTelemetry(ctx context.Context, filter domain.DeviceFilter) (map[string]domain.TelemetryData, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]domain.TelemetryData), args.Error(1)
}

func (m *MockRepository) GetTelemetryHistory(ctx context.Context, filter domain.HistoryFilter) ([]domain.TelemetryRecord, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TelemetryRecord), args.Error(1)
}

// ============================================================
// Service Tests
// ============================================================

func TestTelemetryService_StreamAllDevicesWithHealth_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	healthData := []domain.HealthData{
		{
			DeviceSN:  "SN001",
			Project:   "project-a",
			Type:      "basic-model",
			Status:    "online",
			Uptime:    3600,
			Temp:      25.5,
			LastSeen:  time.Now(),
		},
		{
			DeviceSN:  "SN002",
			Project:   "project-a",
			Type:      "aqi-model",
			Status:    "offline",
			Uptime:    7200,
			Temp:      26.0,
			LastSeen:  time.Now().Add(-1 * time.Hour),
		},
	}

	telemetryData := map[string]domain.TelemetryData{
		"SN001": {
			"temperature": 25.5,
			"humidity":    60.0,
		},
		"SN002": {
			"temperature": 26.0,
			"humidity":    65.0,
		},
	}

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(healthData, nil)
	mockRepo.On("GetLatestTelemetry", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(telemetryData, nil)

	// Execute
	result, err := svc.StreamAllDevicesWithHealth(context.Background(), "project-a")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result)) // 2 device types

	basicModels := result["basic-model"]
	assert.Equal(t, 1, len(basicModels))
	assert.Equal(t, "SN001", basicModels[0].Health.DeviceSN)
	assert.Equal(t, "online", basicModels[0].Health.Status)
	assert.Equal(t, 25.5, basicModels[0].Telemetry["temperature"])

	aqiModels := result["aqi-model"]
	assert.Equal(t, 1, len(aqiModels))
	assert.Equal(t, "SN002", aqiModels[0].Health.DeviceSN)
	assert.Equal(t, "offline", aqiModels[0].Health.Status)

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamAllDevicesWithHealth_NoDevices(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{Project: "empty-project"}).Return([]domain.HealthData{}, nil)
	mockRepo.On("GetLatestTelemetry", mock.Anything, domain.DeviceFilter{Project: "empty-project"}).Return(map[string]domain.TelemetryData{}, nil)

	// Execute
	result, err := svc.StreamAllDevicesWithHealth(context.Background(), "empty-project")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result)) // No devices

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamAllDevicesWithHealth_HealthError(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(nil, assert.AnError)

	// Execute
	result, err := svc.StreamAllDevicesWithHealth(context.Background(), "project-a")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError, err)

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamAllDevicesWithHealth_TelemetryError(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	healthData := []domain.HealthData{
		{
			DeviceSN:  "SN001",
			Type:      "basic-model",
			Status:    "online",
		},
	}

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(healthData, nil)
	mockRepo.On("GetLatestTelemetry", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(nil, assert.AnError)

	// Execute
	result, err := svc.StreamAllDevicesWithHealth(context.Background(), "project-a")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamDeviceWithHealth_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	healthData := []domain.HealthData{
		{
			DeviceSN:  "SN001",
			Project:   "project-a",
			Type:      "basic-model",
			Status:    "online",
			Uptime:    3600,
			Temp:      25.5,
			LastSeen:  time.Now(),
		},
	}

	telemetryData := map[string]domain.TelemetryData{
		"SN001": {
			"temperature": 25.5,
			"humidity":    60.0,
		},
	}

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{Project: "project-a", DeviceSN: "SN001"}).Return(healthData, nil)
	mockRepo.On("GetLatestTelemetry", mock.Anything, domain.DeviceFilter{Project: "project-a", DeviceSN: "SN001"}).Return(telemetryData, nil)

	// Execute
	result, err := svc.StreamDeviceWithHealth(context.Background(), "project-a", "SN001")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SN001", result.Health.DeviceSN)
	assert.Equal(t, "online", result.Health.Status)
	assert.Equal(t, 25.5, result.Telemetry["temperature"])

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamDeviceWithHealth_DeviceNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{Project: "project-a", DeviceSN: "UNKNOWN"}).Return([]domain.HealthData{}, nil)

	// Execute
	result, err := svc.StreamDeviceWithHealth(context.Background(), "project-a", "UNKNOWN")

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result) // Device not found returns nil

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamDeviceWithHealth_NoTelemetry(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	healthData := []domain.HealthData{
		{
			DeviceSN:  "SN001",
			Type:      "basic-model",
			Status:    "online",
		},
	}

	mockRepo.On("GetLatestHealth", mock.Anything, domain.DeviceFilter{DeviceSN: "SN001"}).Return(healthData, nil)
	mockRepo.On("GetLatestTelemetry", mock.Anything, domain.DeviceFilter{DeviceSN: "SN001"}).Return(map[string]domain.TelemetryData{}, nil)

	// Execute
	result, err := svc.StreamDeviceWithHealth(context.Background(), "", "SN001")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SN001", result.Health.DeviceSN)
	assert.NotNil(t, result.Telemetry)
	assert.Equal(t, 0, len(result.Telemetry)) // Empty telemetry is OK

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_GetTelemetryHistory_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	records := []domain.TelemetryRecord{
		{
			Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Fields:    map[string]interface{}{"temperature": 25.5, "humidity": 60.0},
		},
		{
			Timestamp: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			Fields:    map[string]interface{}{"temperature": 26.0, "humidity": 61.0},
		},
	}

	filter := domain.HistoryFilter{
		Project:  "project-a",
		DeviceSN: "SN001",
		TimeRange: domain.QueryTimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Stop:  time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		},
		Window: "1h",
	}

	mockRepo.On("GetTelemetryHistory", mock.Anything, filter).Return(records, nil)

	// Execute
	result, err := svc.GetTelemetryHistory(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 25.5, result[0].Fields["temperature"])

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_GetTelemetryHistory_NilResult(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	filter := domain.HistoryFilter{
		DeviceSN: "SN001",
		TimeRange: domain.QueryTimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Stop:  time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	mockRepo.On("GetTelemetryHistory", mock.Anything, filter).Return(nil, nil)

	// Execute
	result, err := svc.GetTelemetryHistory(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result)) // Should return empty slice

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_GetTelemetryHistory_Error(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	svc := telemetry.NewTelemetryService(mockRepo)

	filter := domain.HistoryFilter{
		DeviceSN: "SN001",
		TimeRange: domain.QueryTimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Stop:  time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		},
	}

	mockRepo.On("GetTelemetryHistory", mock.Anything, filter).Return(nil, assert.AnError)

	// Execute
	result, err := svc.GetTelemetryHistory(context.Background(), filter)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, assert.AnError, err)

	mockRepo.AssertExpectations(t)
}
