package location

import "context"

type locationService struct {
	repo LocationRepository
}

func NewLocationService(repo LocationRepository) LocationService {
	return &locationService{repo: repo}
}

func (s *locationService) CreateLocation(ctx context.Context, req *CreateLocationRequest) (*Location, error) {
	location := &Location{
		Name:  req.Name,
		Notes: req.Notes,
	}

	if err := s.repo.Create(ctx, location); err != nil {
		return nil, err
	}

	return location, nil
}

func (s *locationService) GetLocation(ctx context.Context, guid string) (*Location, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *locationService) ListLocations(ctx context.Context, params *ListParams) (*ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *locationService) UpdateLocation(ctx context.Context, guid string, req *UpdateLocationRequest) (*Location, error) {
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
