package device

import (
	"context"
)

type ListParams struct {
	Limit  int
	Page   int
	Order  string
	Sort   string
	Search string
}

type ListResult struct {
	Devices   []*Device
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

type DeviceRepository interface {
	// Create creates a new device
	Create(ctx context.Context, device *Device) error

	// GetByGUID retrieves a device by GUID
	GetByGUID(ctx context.Context, guid string) (*Device, error)

	// GetBySerialNumber retrieves a device by serial number
	GetBySerialNumber(ctx context.Context, serialNumber string) (*Device, error)

	// List retrieves devices with pagination
	List(ctx context.Context, params ListParams) (*ListResult, error)

	// Update updates an existing device
	Update(ctx context.Context, device *Device) error

	// Delete soft-deletes a device (sets deleted_at)
	Delete(ctx context.Context, guid string) error

	// ExistsBySerialNumber checks if serial number already exists
	ExistsBySerialNumber(ctx context.Context, serialNumber string) (bool, error)
}
