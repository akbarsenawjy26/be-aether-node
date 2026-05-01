package location

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
	Locations []*Location
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

type LocationRepository interface {
	Create(ctx context.Context, location *Location) error
	GetByGUID(ctx context.Context, guid string) (*Location, error)
	List(ctx context.Context, params ListParams) (*ListResult, error)
	Update(ctx context.Context, location *Location) error
	Delete(ctx context.Context, guid string) error
	ExistsByName(ctx context.Context, name string) (bool, error)
}
