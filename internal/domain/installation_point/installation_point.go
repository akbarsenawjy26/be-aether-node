package installation_point

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInstallationPointNotFound = errors.New("installation point not found")
)

type InstallationPoint struct {
	GUID         string     `json:"guid"`
	Name         string     `json:"name"`
	DeviceGUID   string     `json:"device_guid"`
	LocationGUID string     `json:"location_guid"`
	Notes        string     `json:"notes"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	DeviceSN     string     `json:"device_sn,omitempty"`
	DeviceAlias  string     `json:"device_alias,omitempty"`
	LocationName string     `json:"location_name,omitempty"`
}

type InstallationPointWithRelations struct {
	InstallationPoint
	DeviceSerialNumber string `json:"device_serial_number,omitempty"`
	DeviceAlias        string `json:"device_alias,omitempty"`
	LocationName       string `json:"location_name,omitempty"`
}

type ListParams struct {
	Limit        int
	Page         int
	Order        string
	Sort         string
	Search       string
	DeviceGUID   string
	LocationGUID string
}

type ListResult struct {
	InstallationPoints []*InstallationPoint
	Total             int64
	Page              int
	Limit             int
	TotalPage         int
}

type CreateInstallationPointRequest struct {
	Name         string
	DeviceGUID   string
	LocationGUID string
	Notes        string
}

type UpdateInstallationPointRequest struct {
	Name         *string
	DeviceGUID   *string
	LocationGUID *string
	Notes        *string
}

type InstallationPointRepository interface {
	Create(ctx context.Context, ip *InstallationPoint) error
	GetByGUID(ctx context.Context, guid string) (*InstallationPoint, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, ip *InstallationPoint) error
	Delete(ctx context.Context, guid string) error
	GetByGUIDWithRelations(ctx context.Context, guid string) (*InstallationPointWithRelations, error)
	ListWithRelations(ctx context.Context, params ListParams) (*ListResult, error)
}

type InstallationPointService interface {
	CreateInstallationPoint(ctx context.Context, req *CreateInstallationPointRequest) (*InstallationPoint, error)
	GetInstallationPoint(ctx context.Context, guid string) (*InstallationPoint, error)
	GetInstallationPointWithRelations(ctx context.Context, guid string) (*InstallationPointWithRelations, error)
	ListInstallationPoints(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateInstallationPoint(ctx context.Context, guid string, req *UpdateInstallationPointRequest) (*InstallationPoint, error)
	DeleteInstallationPoint(ctx context.Context, guid string) error
}
