package installation_point

import "context"

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

type InstallationPointService interface {
	CreateInstallationPoint(ctx context.Context, req *CreateInstallationPointRequest) (*InstallationPoint, error)
	GetInstallationPoint(ctx context.Context, guid string) (*InstallationPoint, error)
	GetInstallationPointWithRelations(ctx context.Context, guid string) (*InstallationPointWithRelations, error)
	ListInstallationPoints(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateInstallationPoint(ctx context.Context, guid string, req *UpdateInstallationPointRequest) (*InstallationPoint, error)
	DeleteInstallationPoint(ctx context.Context, guid string) error
}
