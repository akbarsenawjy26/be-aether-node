package device

import (
	"context"
	domainDevice "aether-node/internal/domain/device"

	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDeviceNotFound = errors.New("device not found")

type deviceRepository struct {
	db *pgxpool.Pool
}

func NewDeviceRepository(db *pgxpool.Pool) domainDevice.DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(ctx context.Context, device *domainDevice.Device) error {
	if device.GUID == "" {
		device.GUID = uuid.New().String()
	}

	query := `
		INSERT INTO devices (guid, type, serial_number, alias, notes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		device.GUID,
		device.Type,
		device.SerialNumber,
		device.Alias,
		device.Notes,
		true,
		now,
		now,
	)

	return err
}

func (r *deviceRepository) GetByGUID(ctx context.Context, guid string) (*domainDevice.Device, error) {
	query := `
		SELECT guid, type, serial_number, alias, notes, is_active, created_at, updated_at, deleted_at
		FROM devices
		WHERE guid = $1 AND deleted_at IS NULL
	`

	device := &domainDevice.Device{}
	err := r.db.QueryRow(ctx, query, guid).Scan(
		&device.GUID,
		&device.Type,
		&device.SerialNumber,
		&device.Alias,
		&device.Notes,
		&device.IsActive,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDeviceNotFound
	}
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (r *deviceRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (*domainDevice.Device, error) {
	query := `
		SELECT guid, type, serial_number, alias, notes, is_active, created_at, updated_at, deleted_at
		FROM devices
		WHERE serial_number = $1 AND deleted_at IS NULL
	`

	device := &domainDevice.Device{}
	err := r.db.QueryRow(ctx, query, serialNumber).Scan(
		&device.GUID,
		&device.Type,
		&device.SerialNumber,
		&device.Alias,
		&device.Notes,
		&device.IsActive,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDeviceNotFound
	}
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (r *deviceRepository) List(ctx context.Context, params domainDevice.ListParams) (*domainDevice.ListResult, error) {
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

	// Count total
	countQuery := `SELECT COUNT(*) FROM devices WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		countQuery += ` AND (serial_number ILIKE $1 OR alias ILIKE $1 OR type ILIKE $1)`
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Get data
	query := `
		SELECT guid, type, serial_number, alias, notes, is_active, created_at, updated_at, deleted_at
		FROM devices
		WHERE deleted_at IS NULL
	`

	if params.Search != "" {
		query += ` AND (serial_number ILIKE $1 OR alias ILIKE $1 OR type ILIKE $1)`
	}

	query += ` ORDER BY ` + params.Order + ` ` + params.Sort
	query += ` LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := make([]*domainDevice.Device, 0)
	for rows.Next() {
		d := &domainDevice.Device{}
		err := rows.Scan(
			&d.GUID,
			&d.Type,
			&d.SerialNumber,
			&d.Alias,
			&d.Notes,
			&d.IsActive,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}

	totalPages := int(total) / params.Limit
	if int(total)%params.Limit > 0 {
		totalPages++
	}

	return &domainDevice.ListResult{
		Devices:   devices,
		Total:     total,
		Page:      params.Page,
		Limit:     params.Limit,
		TotalPage: totalPages,
	}, nil
}

func (r *deviceRepository) Update(ctx context.Context, device *domainDevice.Device) error {
	query := `
		UPDATE devices
		SET type = $2, serial_number = $3, alias = $4, notes = $5, is_active = $6, updated_at = $7
		WHERE guid = $1 AND deleted_at IS NULL
	`

	device.UpdatedAt = time.Now()
	result, err := r.db.Exec(ctx, query,
		device.GUID,
		device.Type,
		device.SerialNumber,
		device.Alias,
		device.Notes,
		device.IsActive,
		device.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrDeviceNotFound
	}

	return nil
}

func (r *deviceRepository) Delete(ctx context.Context, guid string) error {
	query := `
		UPDATE devices
		SET deleted_at = $2, updated_at = $2
		WHERE guid = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, guid, now)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrDeviceNotFound
	}

	return nil
}

func (r *deviceRepository) ExistsBySerialNumber(ctx context.Context, serialNumber string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM devices WHERE serial_number = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRow(ctx, query, serialNumber).Scan(&exists)
	return exists, err
}
