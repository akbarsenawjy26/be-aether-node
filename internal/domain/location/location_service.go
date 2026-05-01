package location

import "context"

type CreateLocationRequest struct {
	Name  string
	Notes string
}

type UpdateLocationRequest struct {
	Name  *string
	Notes *string
}

type LocationService interface {
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (*Location, error)
	GetLocation(ctx context.Context, guid string) (*Location, error)
	ListLocations(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateLocation(ctx context.Context, guid string, req *UpdateLocationRequest) (*Location, error)
	DeleteLocation(ctx context.Context, guid string) error
}
