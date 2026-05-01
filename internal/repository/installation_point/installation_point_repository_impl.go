package installation_point

import (
	"context"
	domainIP "aether-node/internal/domain/installation_point"

	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInstallationPointNotFound = errors.New("installation point not found")

type installationPointRepository struct {
	db *pgxpool.Pool
}

func NewInstallationPointRepository(db *pgxpool.Pool) domainIP.InstallationPointRepository {
	return &installationPointRepository{db: db}
}

func (r *installationPointRepository) Create(ctx context.Context, ip *domainIP.InstallationPoint) error {
	if ip.GUID == "" {
		ip.GUID = uuid.New().String()
	}

	query := `
		INSERT INTO installation_points (guid, name, device_guid, location_guid, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		ip.GUID,
		ip.Name,
		ip.DeviceGUID,
		ip.LocationGUID,
		ip.Notes,
		now,
		now,
	)

	return err
}

func (r *installationPointRepository) GetByGUID(ctx context.Context, guid string) (*domainIP.InstallationPoint, error) {
	query := `
		SELECT guid, name, device_guid, location_guid, notes, created_at, updated_at, deleted_at
		FROM installation_points
		WHERE guid = $1 AND deleted_at IS NULL
	`

	ip := &domainIP.InstallationPoint{}
	err := r.db.QueryRow(ctx, query, guid).Scan(
		&ip.GUID,
		&ip.Name,
		&ip.DeviceGUID,
		&ip.LocationGUID,
		&ip.Notes,
		&ip.CreatedAt,
		&ip.UpdatedAt,
		&ip.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInstallationPointNotFound
	}
	if err != nil {
		return nil, err
	}

	return ip, nil
}

func (r *installationPointRepository) List(ctx context.Context, params domainIP.ListParams) (*domainIP.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Order == "" {
		params.Order = "created_at"
	}
	if params.Sort == "" {
		params.Sort = "DESC"
	}

	offset := (params.Page - 1) * params.Limit

	// Build query with filters
	whereClause := "WHERE ip.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		whereClause += ` AND (ip.name ILIKE $` + string(rune('0'+argIdx)) + ` OR ip.notes ILIKE $` + string(rune('0'+argIdx)) + `)`
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	if params.DeviceGUID != "" {
		whereClause += ` AND ip.device_guid = $` + string(rune('0'+argIdx))
		args = append(args, params.DeviceGUID)
		argIdx++
	}

	if params.LocationGUID != "" {
		whereClause += ` AND ip.location_guid = $` + string(rune('0'+argIdx))
		args = append(args, params.LocationGUID)
		argIdx++
	}

	// Count total
	countQuery := `
		SELECT COUNT(*) 
		FROM installation_points ip
		` + whereClause

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get data
	query := `
		SELECT ip.guid, ip.name, ip.device_guid, ip.location_guid, ip.notes, ip.created_at, ip.updated_at, ip.deleted_at
		FROM installation_points ip
		` + whereClause + `
		ORDER BY ip.` + params.Order + ` ` + params.Sort + `
		LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]*domainIP.InstallationPoint, 0)
	for rows.Next() {
		ip := &domainIP.InstallationPoint{}
		err := rows.Scan(
			&ip.GUID,
			&ip.Name,
			&ip.DeviceGUID,
			&ip.LocationGUID,
			&ip.Notes,
			&ip.CreatedAt,
			&ip.UpdatedAt,
			&ip.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		ips = append(ips, ip)
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
	query := `
		UPDATE installation_points
		SET name = $2, device_guid = $3, location_guid = $4, notes = $5, updated_at = $6
		WHERE guid = $1 AND deleted_at IS NULL
	`

	ip.UpdatedAt = time.Now()
	result, err := r.db.Exec(ctx, query,
		ip.GUID,
		ip.Name,
		ip.DeviceGUID,
		ip.LocationGUID,
		ip.Notes,
		ip.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrInstallationPointNotFound
	}

	return nil
}

func (r *installationPointRepository) Delete(ctx context.Context, guid string) error {
	query := `
		UPDATE installation_points
		SET deleted_at = $2, updated_at = $2
		WHERE guid = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, guid, now)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrInstallationPointNotFound
	}

	return nil
}

func (r *installationPointRepository) GetByGUIDWithRelations(ctx context.Context, guid string) (*domainIP.InstallationPointWithRelations, error) {
	query := `
		SELECT 
			ip.guid, ip.name, ip.device_guid, ip.location_guid, ip.notes, ip.created_at, ip.updated_at, ip.deleted_at,
			d.serial_number, d.alias, l.name
		FROM installation_points ip
		LEFT JOIN devices d ON ip.device_guid = d.guid
		LEFT JOIN locations l ON ip.location_guid = l.guid
		WHERE ip.guid = $1 AND ip.deleted_at IS NULL
	`

	ip := &domainIP.InstallationPointWithRelations{}
	err := r.db.QueryRow(ctx, query, guid).Scan(
		&ip.GUID,
		&ip.Name,
		&ip.DeviceGUID,
		&ip.LocationGUID,
		&ip.Notes,
		&ip.CreatedAt,
		&ip.UpdatedAt,
		&ip.DeletedAt,
		&ip.DeviceSerialNumber,
		&ip.DeviceAlias,
		&ip.LocationName,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInstallationPointNotFound
	}
	if err != nil {
		return nil, err
	}

	return ip, nil
}

func (r *installationPointRepository) ListWithRelations(ctx context.Context, params domainIP.ListParams) (*domainIP.ListResult, error) {
	// For simplicity, we reuse the List method and enhance in service layer
	// In production, you might want a dedicated query with JOINs
	return r.List(ctx, params)
}
