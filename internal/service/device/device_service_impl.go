package device

import (
	"context"
	"errors"
)

var (
	ErrDeviceSerialNumberExists = errors.New("serial number already exists")
)

type deviceService struct {
	repo DeviceRepository
}

func NewDeviceService(repo DeviceRepository) DeviceService {
	return &deviceService{repo: repo}
}

func (s *deviceService) CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*Device, error) {
	// Check if serial number already exists
	exists, err := s.repo.ExistsBySerialNumber(ctx, req.SerialNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDeviceSerialNumberExists
	}

	device := &Device{
		Type:         req.Type,
		SerialNumber: req.SerialNumber,
		Alias:        req.Alias,
		Notes:        req.Notes,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *deviceService) GetDevice(ctx context.Context, guid string) (*Device, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *deviceService) GetDeviceBySerialNumber(ctx context.Context, serialNumber string) (*Device, error) {
	return s.repo.GetBySerialNumber(ctx, serialNumber)
}

func (s *deviceService) ListDevices(ctx context.Context, params *ListParams) (*ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *deviceService) UpdateDevice(ctx context.Context, guid string, req *UpdateDeviceRequest) (*Device, error) {
	device, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	// Check if new serial number already exists
	if req.SerialNumber != nil && *req.SerialNumber != device.SerialNumber {
		exists, err := s.repo.ExistsBySerialNumber(ctx, *req.SerialNumber)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrDeviceSerialNumberExists
		}
		device.SerialNumber = *req.SerialNumber
	}

	if req.Type != nil {
		device.Type = *req.Type
	}

	if req.Alias != nil {
		device.Alias = *req.Alias
	}

	if req.Notes != nil {
		device.Notes = *req.Notes
	}

	if req.IsActive != nil {
		device.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *deviceService) DeleteDevice(ctx context.Context, guid string) error {
	return s.repo.Delete(ctx, guid)
}
