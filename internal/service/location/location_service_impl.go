package location

import (
	"context"

	domainLocation "aether-node/internal/domain/location"
)

type locationService struct {
	repo domainLocation.LocationRepository
}

func NewLocationService(repo domainLocation.LocationRepository) domainLocation.LocationService {
	return &locationService{repo: repo}
}

func (s *locationService) CreateLocation(ctx context.Context, req *domainLocation.CreateLocationRequest) (*domainLocation.Location, error) {
	location := &domainLocation.Location{
		Name:  req.Name,
		Notes: req.Notes,
	}

	if err := s.repo.Create(ctx, location); err != nil {
		return nil, err
	}

	return location, nil
}

func (s *locationService) GetLocation(ctx context.Context, guid string) (*domainLocation.Location, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *locationService) ListLocations(ctx context.Context, params *domainLocation.ListParams) (*domainLocation.ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *locationService) UpdateLocation(ctx context.Context, guid string, req *domainLocation.UpdateLocationRequest) (*domainLocation.Location, error) {
	location, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		location.Name = *req.Name
	}

	if req.Notes != nil {
		location.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, location); err != nil {
		return nil, err
	}

	return location, nil
}

func (s *locationService) DeleteLocation(ctx context.Context, guid string) error {
	return s.repo.Delete(ctx, guid)
}
