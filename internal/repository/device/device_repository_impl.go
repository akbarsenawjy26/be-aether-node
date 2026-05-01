package device

import (
	"context"
	"errors"
	"time"

	"aether-node/internal/db"
	domainDevice "aether-node/internal/domain/device"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrDeviceNotFound = errors.New("device not found")

type deviceRepository struct {
	db *db.Queries
}

func NewDeviceRepository(queries *db.Queries) domainDevice.DeviceRepository {
	return &deviceRepository{db: queries}
}

func (r *deviceRepository) Create(ctx context.Context, device *domainDevice.Device) error {
	if device.GUID == "" {
		device.GUID = uuid.New().String()
	}

	guid, _ := uuid.Parse(device.GUID)
	now := time.Now()

	params := db.CreateDeviceParams{
		Guid:         db.NewUUID(guid),
		Type:         device.Type,
		SerialNumber: device.SerialNumber,
		IsActive:     device.IsActive,
		CreatedAt:    db.NewTimestamptz(now),
		UpdatedAt:    db.NewTimestamptz(now),
	}

	if device.Alias != "" {
		params.Alias = db.NewText(device.Alias)
	}
	if device.Notes != "" {
		params.Notes = db.NewText(device.Notes)
	}

	return r.db.CreateDevice(ctx, params)
}

func (r *deviceRepository) GetByGUID(ctx context.Context, guid string) (*domainDevice.Device, error) {
	id, err := uuid.Parse(guid)
	if err != nil {
		return nil, ErrDeviceNotFound
	}

	dbDevice, err := r.db.GetDeviceByGUID(ctx, db.NewUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}

	return db.DeviceFromDB(&dbDevice), nil
}

func (r *deviceRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (*domainDevice.Device, error) {
	dbDevice, err := r.db.GetDeviceBySerialNumber(ctx, serialNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, err
	}

	return db.DeviceFromDB(&dbDevice), nil
}

func (r *deviceRepository) List(ctx context.Context, params domainDevice.ListParams) (*domainDevice.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	search := "%" + params.Search + "%"
	offset := (params.Page - 1) * params.Limit

	dbDevices, err := r.db.ListDevices(ctx, db.ListDevicesParams{
		SerialNumber: search,
		Limit:        int32(params.Limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, err
	}

	total, err := r.db.CountDevices(ctx)
	if err != nil {
		return nil, err
	}

	devices := make([]*domainDevice.Device, 0, len(dbDevices))
	for i := range dbDevices {
		devices = append(devices, db.DeviceFromDB(&dbDevices[i]))
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
	guid, _ := uuid.Parse(device.GUID)
	now := time.Now()

	params := db.UpdateDeviceParams{
		Guid:         db.NewUUID(guid),
		Type:         device.Type,
		SerialNumber: device.SerialNumber,
		IsActive:     device.IsActive,
		UpdatedAt:    db.NewTimestamptz(now),
	}

	if device.Alias != "" {
		params.Alias = db.NewText(device.Alias)
	}
	if device.Notes != "" {
		params.Notes = db.NewText(device.Notes)
	}

	err := r.db.UpdateDevice(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDeviceNotFound
		}
		return err
	}
	return nil
}

func (r *deviceRepository) Delete(ctx context.Context, guid string) error {
	id, err := uuid.Parse(guid)
	if err != nil {
		return ErrDeviceNotFound
	}

	err = r.db.DeleteDevice(ctx, db.DeleteDeviceParams{
		Guid:      db.NewUUID(id),
		DeletedAt: db.NewTimestamptz(time.Now()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrDeviceNotFound
		}
		return err
	}
	return nil
}

func (r *deviceRepository) ExistsBySerialNumber(ctx context.Context, serialNumber string) (bool, error) {
	exists, err := r.db.ExistsDeviceBySerialNumber(ctx, serialNumber)
	return bool(exists), err
}
