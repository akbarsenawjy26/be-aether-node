package device

import "context"

type CreateDeviceRequest struct {
	Type         string
	SerialNumber string
	Alias        string
	Notes        string
}

type UpdateDeviceRequest struct {
	Type         *string
	SerialNumber *string
	Alias        *string
	Notes        *string
	IsActive     *bool
}

type DeviceService interface {
	// CreateDevice creates a new device
	CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*Device, error)

	// GetDevice retrieves a device by GUID
	GetDevice(ctx context.Context, guid string) (*Device, error)

	// GetDeviceBySerialNumber retrieves a device by serial number
	GetDeviceBySerialNumber(ctx context.Context, serialNumber string) (*Device, error)

	// ListDevices retrieves devices with pagination
	ListDevices(ctx context.Context, params *ListParams) (*ListResult, error)

	// UpdateDevice updates an existing device
	UpdateDevice(ctx context.Context, guid string, req *UpdateDeviceRequest) (*Device, error)

	// DeleteDevice soft-deletes a device
	DeleteDevice(ctx context.Context, guid string) error
}
