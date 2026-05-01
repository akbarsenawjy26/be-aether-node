package installation_point

import (
	"context"
	"errors"
	"time"

	"aether-node/internal/db"
	domainIP "aether-node/internal/domain/installation_point"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrInstallationPointNotFound = errors.New("installation point not found")

type installationPointRepository struct {
	db *db.Queries
}

func NewInstallationPointRepository(queries *db.Queries) domainIP.InstallationPointRepository {
	return &installationPointRepository{db: queries}
}

func (r *installationPointRepository) Create(ctx context.Context, ip *domainIP.InstallationPoint) error {
	if ip.GUID == "" {
		ip.GUID = uuid.New().String()
	}

	guid, _ := uuid.Parse(ip.GUID)
	deviceGUID, _ := uuid.Parse(ip.DeviceGUID)
	locationGUID, _ := uuid.Parse(ip.LocationGUID)
	now := time.Now()

	params := db.CreateInstallationPointParams{
		Guid:         db.NewUUID(guid),
		Name:         ip.Name,
		DeviceGuid:   db.NewUUID(deviceGUID),
		LocationGuid: db.NewUUID(locationGUID),
		CreatedAt:    db.NewTimestamptz(now),
		UpdatedAt:    db.NewTimestamptz(now),
	}

	if ip.Notes != "" {
		params.Notes = db.NewText(ip.Notes)
	}

	return r.db.CreateInstallationPoint(ctx, params)
}

func (r *installationPointRepository) GetByGUID(ctx context.Context, guid string) (*domainIP.InstallationPoint, error) {
	id, err := uuid.Parse(guid)
	if err != nil {
		return nil, ErrInstallationPointNotFound
	}

	dbIP, err := r.db.GetInstallationPointByGUID(ctx, db.NewUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstallationPointNotFound
		}
		return nil, err
	}

	return db.InstallationPointFromDB(&dbIP), nil
}

func (r *installationPointRepository) List(ctx context.Context, params domainIP.ListParams) (*domainIP.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	var search string
	if params.Search != "" {
		search = "%" + params.Search + "%"
	}

	var deviceGUID pgtype.UUID
	if params.DeviceGUID != "" {
		id, _ := uuid.Parse(params.DeviceGUID)
		deviceGUID = db.NewUUID(id)
	}

	var locationGUID pgtype.UUID
	if params.LocationGUID != "" {
		id, _ := uuid.Parse(params.LocationGUID)
		locationGUID = db.NewUUID(id)
	}

	offset := (params.Page - 1) * params.Limit

	dbIPs, err := r.db.ListInstallationPoints(ctx, db.ListInstallationPointsParams{
		Column1: search,
		Column2: deviceGUID,
		Column3: locationGUID,
		Limit:   int32(params.Limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, err
	}

	total, err := r.db.CountInstallationPoints(ctx)
	if err != nil {
		return nil, err
	}

	ips := make([]*domainIP.InstallationPoint, 0, len(dbIPs))
	for i := range dbIPs {
		ips = append(ips, db.InstallationPointFromDB(&dbIPs[i]))
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &domainIP.ListResult{
		InstallationPoints: ips,
		Total:             total,
		Page:              params.Page,
		Limit:             params.Limit,
		TotalPage:         totalPages,
	}, nil
}

func (r *installationPointRepository) Update(ctx context.Context, ip *domainIP.InstallationPoint) error {
	guid, _ := uuid.Parse(ip.GUID)
	deviceGUID, _ := uuid.Parse(ip.DeviceGUID)
	locationGUID, _ := uuid.Parse(ip.LocationGUID)
	now := time.Now()

	params := db.UpdateInstallationPointParams{
		Guid:         db.NewUUID(guid),
		Name:         ip.Name,
		DeviceGuid:   db.NewUUID(deviceGUID),
		LocationGuid: db.NewUUID(locationGUID),
		UpdatedAt:    db.NewTimestamptz(now),
	}

	if ip.Notes != "" {
		params.Notes = db.NewText(ip.Notes)
	}

	err := r.db.UpdateInstallationPoint(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInstallationPointNotFound
		}
		return err
	}
	return nil
}

func (r *installationPointRepository) Delete(ctx context.Context, guid string) error {
	id, err := uuid.Parse(guid)
	if err != nil {
		return ErrInstallationPointNotFound
	}

	err = r.db.DeleteInstallationPoint(ctx, db.DeleteInstallationPointParams{
		Guid:      db.NewUUID(id),
		DeletedAt: db.NewTimestamptz(time.Now()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInstallationPointNotFound
		}
		return err
	}
	return nil
}

func (r *installationPointRepository) GetByGUIDWithRelations(ctx context.Context, guid string) (*domainIP.InstallationPointWithRelations, error) {
	id, err := uuid.Parse(guid)
	if err != nil {
		return nil, ErrInstallationPointNotFound
	}

	row, err := r.db.GetInstallationPointWithRelations(ctx, db.NewUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInstallationPointNotFound
		}
		return nil, err
	}

	ip := &domainIP.InstallationPoint{
		Name: row.Name,
	}
	if row.Guid.Valid {
		ip.GUID = uuid.UUID(row.Guid.Bytes).String()
	}
	if row.DeviceGuid.Valid {
		ip.DeviceGUID = uuid.UUID(row.DeviceGuid.Bytes).String()
	}
	if row.LocationGuid.Valid {
		ip.LocationGUID = uuid.UUID(row.LocationGuid.Bytes).String()
	}
	if row.Notes.Valid {
		ip.Notes = row.Notes.String
	}
	if row.CreatedAt.Valid {
		ip.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		ip.UpdatedAt = row.UpdatedAt.Time
	}
	if row.DeletedAt.Valid {
		ip.DeletedAt = &row.DeletedAt.Time
	}

	result := &domainIP.InstallationPointWithRelations{
		InstallationPoint: *ip,
	}
	if row.SerialNumber.Valid {
		result.DeviceSerialNumber = row.SerialNumber.String
	}
	if row.Alias.Valid {
		result.DeviceAlias = row.Alias.String
	}
	if row.Name_2.Valid {
		result.LocationName = row.Name_2.String
	}

	return result, nil
}

func (r *installationPointRepository) ListWithRelations(ctx context.Context, params domainIP.ListParams) (*domainIP.ListResult, error) {
	return r.List(ctx, params)
}
