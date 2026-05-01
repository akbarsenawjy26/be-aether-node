package location

import (
	"context"
	"errors"
	"time"
)

var (
	ErrLocationNotFound = errors.New("location not found")
)

type Location struct {
	GUID      string     `json:"guid"`
	Name      string     `json:"name"`
	Notes     string     `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type ListParams struct {
	Limit  int
	Page   int
	Order  string
	Sort   string
	Search string
}

type ListResult struct {
	Locations []*Location
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

type CreateLocationRequest struct {
	Name  string
	Notes string
}

type UpdateLocationRequest struct {
	Name  *string
	Notes *string
}

type LocationRepository interface {
	Create(ctx context.Context, location *Location) error
	GetByGUID(ctx context.Context, guid string) (*Location, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, location *Location) error
	Delete(ctx context.Context, guid string) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type LocationService interface {
	CreateLocation(ctx context.Context, req *CreateLocationRequest) (*Location, error)
	GetLocation(ctx context.Context, guid string) (*Location, error)
	ListLocations(ctx context.Context, params *ListParams) (*ListResult, error)
	UpdateLocation(ctx context.Context, guid string, req *UpdateLocationRequest) (*Location, error)
	DeleteLocation(ctx context.Context, guid string) error
}
