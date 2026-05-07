package threshold

import (
	"context"
	"errors"
	"fmt"

	"aether-node/internal/db"
	"aether-node/internal/domain/threshold"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	// "github.com/jackc/pgx/v5/pgtype"
)

type thresholdRepositoryImpl struct {
	db *db.Queries
}

func NewThresholdRepository(queries *db.Queries) threshold.ThresholdRepository {
	return &thresholdRepositoryImpl{db: queries}
}

func (r *thresholdRepositoryImpl) Create(ctx context.Context, t *threshold.Threshold) error {
	deviceUUID, err := uuid.Parse(t.DeviceGUID)
	if err != nil {
		return fmt.Errorf("invalid device guid: %w", err)
	}

	arg := db.CreateThresholdParams{
		DeviceGuid:    db.NewUUID(deviceUUID),
		ParameterName: t.ParameterName,
		MinValue:      db.NewFloat8(t.MinValue),
		MaxValue:      db.NewFloat8(t.MaxValue),
		Severity:      db.AlarmSeverity(t.Severity),
		IsActive:      t.IsActive,
	}

	res, err := r.db.CreateThreshold(ctx, arg)
	if err != nil {
		return err
	}

	t.GUID = uuid.UUID(res.Guid.Bytes).String()
	t.CreatedAt = res.CreatedAt.Time
	t.UpdatedAt = res.UpdatedAt.Time
	return nil
}

func (r *thresholdRepositoryImpl) GetByGUID(ctx context.Context, guidStr string) (*threshold.Threshold, error) {
	u, err := uuid.Parse(guidStr)
	if err != nil {
		return nil, threshold.ErrThresholdNotFound
	}

	res, err := r.db.GetThresholdByGUID(ctx, db.NewUUID(u))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, threshold.ErrThresholdNotFound
		}
		return nil, err
	}

	return db.ThresholdFromDB(&res), nil
}

func (r *thresholdRepositoryImpl) ListByDevice(ctx context.Context, deviceGUID string) ([]*threshold.Threshold, error) {
	u, err := uuid.Parse(deviceGUID)
	if err != nil {
		return nil, nil
	}

	res, err := r.db.ListThresholdsByDevice(ctx, db.NewUUID(u))
	if err != nil {
		return nil, err
	}

	var results []*threshold.Threshold
	for i := range res {
		results = append(results, db.ThresholdFromDB(&res[i]))
	}
	return results, nil
}

func (r *thresholdRepositoryImpl) Update(ctx context.Context, t *threshold.Threshold) error {
	u, err := uuid.Parse(t.GUID)
	if err != nil {
		return threshold.ErrThresholdNotFound
	}

	arg := db.UpdateThresholdParams{
		Guid:          db.NewUUID(u),
		ParameterName: t.ParameterName,
		MinValue:      db.NewFloat8(t.MinValue),
		MaxValue:      db.NewFloat8(t.MaxValue),
		Severity:      db.AlarmSeverity(t.Severity),
		IsActive:      t.IsActive,
	}

	res, err := r.db.UpdateThreshold(ctx, arg)
	if err != nil {
		return err
	}

	t.UpdatedAt = res.UpdatedAt.Time
	return nil
}

func (r *thresholdRepositoryImpl) Delete(ctx context.Context, guidStr string) error {
	u, err := uuid.Parse(guidStr)
	if err != nil {
		return threshold.ErrThresholdNotFound
	}

	return r.db.DeleteThreshold(ctx, db.NewUUID(u))
}
