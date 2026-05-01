package device

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDeviceNotFound = errors.New("device not found")
	ErrDeviceSerialNumberExists = errors.New("serial number already exists")
)

type Device struct {
	GUID         string     `json:"guid"`
	Type         string     `json:"type"`
	SerialNumber string     `json:"serial_number"`
	Alias        string     `json:"alias"`
	Notes        string     `json:"notes"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

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

type DeviceRepository interface {
	Create(ctx context.Context, device *Device) error
	GetByGUID(ctx context.Context, guid string) (*Device, error)
	GetBySerialNumber(ctx context.Context, serialNumber string) (*Device, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, device *Device) error
	Delete(ctx context.Context, guid string) error
	ExistsBySerialNumber(ctx context.Context, serialNumber string) (bool, error)
}

type DeviceService interface {
	CreateDevice(ctx context.Context, req *CreateDeviceRequest) (*Device, error)
	GetDevice(ctx context.Context, guid string) (*Device, error)
	GetDeviceBySerialNumber(ctx context.Context, serialNumber string) (*Device, error)
	ListDevices(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateDevice(ctx context.Context, guid string, req *UpdateDeviceRequest) (*Device, error)
	DeleteDevice(ctx context.Context, guid string) error
}
