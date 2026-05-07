package threshold

import (
	"context"
	"errors"
	"time"
)

var (
	ErrThresholdNotFound = errors.New("threshold not found")
)

type Threshold struct {
	GUID          string     `json:"guid"`
	DeviceGUID    string     `json:"device_guid"`
	ParameterName string     `json:"parameter_name"`
	MinValue      *float64   `json:"min_value"`
	MaxValue      *float64   `json:"max_value"`
	Severity      string     `json:"severity"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type CreateThresholdRequest struct {
	DeviceGUID    string   `json:"device_guid" validate:"required"`
	ParameterName string   `json:"parameter_name" validate:"required"`
	MinValue      *float64 `json:"min_value"`
	MaxValue      *float64 `json:"max_value"`
	Severity      string   `json:"severity" validate:"required"`
	IsActive      bool     `json:"is_active"`
}

type UpdateThresholdRequest struct {
	ParameterName *string  `json:"parameter_name"`
	MinValue      *float64 `json:"min_value"`
	MaxValue      *float64 `json:"max_value"`
	Severity      *string  `json:"severity"`
	IsActive      *bool    `json:"is_active"`
}

type ThresholdRepository interface {
	Create(ctx context.Context, t *Threshold) error
	GetByGUID(ctx context.Context, guid string) (*Threshold, error)
	ListByDevice(ctx context.Context, deviceGUID string) ([]*Threshold, error)
	Update(ctx context.Context, t *Threshold) error
	Delete(ctx context.Context, guid string) error
}

type ThresholdService interface {
	CreateThreshold(ctx context.Context, req *CreateThresholdRequest) (*Threshold, error)
	GetThreshold(ctx context.Context, guid string) (*Threshold, error)
	ListThresholdsByDevice(ctx context.Context, deviceGUID string) ([]*Threshold, error)
	UpdateThreshold(ctx context.Context, guid string, req *UpdateThresholdRequest) (*Threshold, error)
	DeleteThreshold(ctx context.Context, guid string) error
}
