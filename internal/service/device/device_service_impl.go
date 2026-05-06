package device

import (
	"context"

	domainDevice "aether-node/internal/domain/device"
)

type deviceService struct {
	repo     domainDevice.DeviceRepository
	onUpdate func() // Callback to invalidate external caches
}

func NewDeviceService(repo domainDevice.DeviceRepository, onUpdate func()) domainDevice.DeviceService {
	return &deviceService{
		repo:     repo,
		onUpdate: onUpdate,
	}
}

func (s *deviceService) CreateDevice(ctx context.Context, req *domainDevice.CreateDeviceRequest) (*domainDevice.Device, error) {
	exists, err := s.repo.ExistsBySerialNumber(ctx, req.SerialNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainDevice.ErrDeviceSerialNumberExists
	}

	device := &domainDevice.Device{
		Type:         req.Type,
		SerialNumber: req.SerialNumber,
		Alias:        req.Alias,
		Notes:        req.Notes,
		IsActive:     true,
	}

	if err := s.repo.Create(ctx, device); err != nil {
		return nil, err
	}

	if s.onUpdate != nil {
		s.onUpdate()
	}

	return device, nil
}

func (s *deviceService) GetDevice(ctx context.Context, guid string) (*domainDevice.Device, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *deviceService) GetDeviceBySerialNumber(ctx context.Context, serialNumber string) (*domainDevice.Device, error) {
	return s.repo.GetBySerialNumber(ctx, serialNumber)
}

func (s *deviceService) ListDevices(ctx context.Context, params *domainDevice.ListParams) (*domainDevice.ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *deviceService) UpdateDevice(ctx context.Context, guid string, req *domainDevice.UpdateDeviceRequest) (*domainDevice.Device, error) {
	device, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	if req.SerialNumber != nil && *req.SerialNumber != device.SerialNumber {
		exists, err := s.repo.ExistsBySerialNumber(ctx, *req.SerialNumber)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domainDevice.ErrDeviceSerialNumberExists
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

	if s.onUpdate != nil {
		s.onUpdate()
	}

	return device, nil
}

func (s *deviceService) DeleteDevice(ctx context.Context, guid string) error {
	if err := s.repo.Delete(ctx, guid); err != nil {
		return err
	}

	if s.onUpdate != nil {
		s.onUpdate()
	}

	return nil
}
