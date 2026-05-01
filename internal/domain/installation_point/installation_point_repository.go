package installation_point

import (
	"context"
)

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

type InstallationPointRepository interface {
	Create(ctx context.Context, ip *InstallationPoint) error
	GetByGUID(ctx context.Context, guid string) (*InstallationPoint, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, ip *InstallationPoint) error
	Delete(ctx context.Context, guid string) error
	GetByGUIDWithRelations(ctx context.Context, guid string) (*InstallationPointWithRelations, error)
	ListWithRelations(ctx context.Context, params ListParams) (*ListResult, error)
}
