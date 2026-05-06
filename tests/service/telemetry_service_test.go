package service_test

import (
	"context"
	"testing"
	"time"

	domainDevice "aether-node/internal/domain/device"
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

func (m *MockRepository) GetLatestHealthSSE(ctx context.Context, filter domain.DeviceFilter) ([]domain.HealthData, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.HealthData), args.Error(1)
}

func (m *MockRepository) GetLatestTelemetrySSE(ctx context.Context, filter domain.DeviceFilter) (map[string]domain.TelemetryData, error) {
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

type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *domainDevice.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetByGUID(ctx context.Context, guid string) (*domainDevice.Device, error) {
	args := m.Called(ctx, guid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainDevice.Device), args.Error(1)
}

func (m *MockDeviceRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (*domainDevice.Device, error) {
	args := m.Called(ctx, serialNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainDevice.Device), args.Error(1)
}

func (m *MockDeviceRepository) List(ctx context.Context, params domainDevice.ListParams) (*domainDevice.ListResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainDevice.ListResult), args.Error(1)
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *domainDevice.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Delete(ctx context.Context, guid string) error {
	args := m.Called(ctx, guid)
	return args.Error(0)
}

func (m *MockDeviceRepository) ExistsBySerialNumber(ctx context.Context, serialNumber string) (bool, error) {
	args := m.Called(ctx, serialNumber)
	return args.Bool(0), args.Error(1)
}

// ============================================================
// Service Tests
// ============================================================

func TestTelemetryService_StreamAllDevicesWithHealth_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockDeviceRepo := new(MockDeviceRepository)
	svc := telemetry.NewTelemetryService(mockRepo, mockDeviceRepo)

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

	mockDeviceRepo.On("List", mock.Anything, mock.Anything).Return(&domainDevice.ListResult{
		Devices: []*domainDevice.Device{
			{SerialNumber: "SN001", Alias: "Device One"},
			{SerialNumber: "SN002", Alias: "Device Two"},
		},
	}, nil)

	mockRepo.On("GetLatestHealthSSE", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(healthData, nil)
	mockRepo.On("GetLatestTelemetrySSE", mock.Anything, domain.DeviceFilter{Project: "project-a"}).Return(telemetryData, nil)

	// Execute
	result, err := svc.StreamAllDevicesWithHealth(context.Background(), "project-a")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result)) // 2 device types

	basicModels := result["basic-model"]
	assert.Equal(t, 1, len(basicModels))
	assert.Equal(t, "SN001", basicModels[0].Health.DeviceSN)
	assert.Equal(t, "Device One", basicModels[0].Health.DeviceName)
	assert.Equal(t, "online", basicModels[0].Health.Status)
	assert.Equal(t, 25.5, basicModels[0].Telemetry["temperature"])

	aqiModels := result["aqi-model"]
	assert.Equal(t, 1, len(aqiModels))
	assert.Equal(t, "SN002", aqiModels[0].Health.DeviceSN)
	assert.Equal(t, "Device Two", aqiModels[0].Health.DeviceName)
	assert.Equal(t, "offline", aqiModels[0].Health.Status)

	mockRepo.AssertExpectations(t)
}

func TestTelemetryService_StreamDeviceWithHealth_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockRepository)
	mockDeviceRepo := new(MockDeviceRepository)
	svc := telemetry.NewTelemetryService(mockRepo, mockDeviceRepo)

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

	mockDeviceRepo.On("GetBySerialNumber", mock.Anything, "SN001").Return(&domainDevice.Device{
		SerialNumber: "SN001",
		Alias:        "Device One",
	}, nil)

	mockRepo.On("GetLatestHealthSSE", mock.Anything, domain.DeviceFilter{Project: "project-a", DeviceSN: "SN001"}).Return(healthData, nil)
	mockRepo.On("GetLatestTelemetrySSE", mock.Anything, domain.DeviceFilter{Project: "project-a", DeviceSN: "SN001"}).Return(telemetryData, nil)

	// Execute
	result, err := svc.StreamDeviceWithHealth(context.Background(), "project-a", "SN001")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "SN001", result.Health.DeviceSN)
	assert.Equal(t, "Device One", result.Health.DeviceName)
	assert.Equal(t, "online", result.Health.Status)
	assert.Equal(t, 25.5, result.Telemetry["temperature"])

	mockRepo.AssertExpectations(t)
}
