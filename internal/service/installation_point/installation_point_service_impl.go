package installation_point

import "context"

type installationPointService struct {
	repo InstallationPointRepository
}

func NewInstallationPointService(repo InstallationPointRepository) InstallationPointService {
	return &installationPointService{repo: repo}
}

func (s *installationPointService) CreateInstallationPoint(ctx context.Context, req *CreateInstallationPointRequest) (*InstallationPoint, error) {
	ip := &InstallationPoint{
		Name:         req.Name,
		DeviceGUID:   req.DeviceGUID,
		LocationGUID: req.LocationGUID,
		Notes:        req.Notes,
	}

	if err := s.repo.Create(ctx, ip); err != nil {
		return nil, err
	}

	return ip, nil
}

func (s *installationPointService) GetInstallationPoint(ctx context.Context, guid string) (*InstallationPoint, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *installationPointService) GetInstallationPointWithRelations(ctx context.Context, guid string) (*InstallationPointWithRelations, error) {
	return s.repo.GetByGUIDWithRelations(ctx, guid)
}

func (s *installationPointService) ListInstallationPoints(ctx context.Context, params *ListParams) (*ListResult, error) {
	return s.repo.List(ctx, *params)
}

func (s *installationPointService) UpdateInstallationPoint(ctx context.Context, guid string, req *UpdateInstallationPointRequest) (*InstallationPoint, error) {
	ip, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		ip.Name = *req.Name
	}

	if req.DeviceGUID != nil {
		ip.DeviceGUID = *req.DeviceGUID
	}

	if req.LocationGUID != nil {
		ip.LocationGUID = *req.LocationGUID
	}

	if req.Notes != nil {
		ip.Notes = *req.Notes
	}

	if err := s.repo.Update(ctx, ip); err != nil {
		return nil, err
	}

	return ip, nil
}

func (s *installationPointService) DeleteInstallationPoint(ctx context.Context, guid string) error {
	return s.repo.Delete(ctx, guid)
}
