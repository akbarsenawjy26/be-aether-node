package alarm

import (
	"context"
	"errors"
	"fmt"

	"aether-node/internal/db"
	"aether-node/internal/domain/alarm"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type alarmRepositoryImpl struct {
	db *db.Queries
}

func NewAlarmRepository(queries *db.Queries) alarm.AlarmRepository {
	return &alarmRepositoryImpl{db: queries}
}

func (r *alarmRepositoryImpl) Create(ctx context.Context, a *alarm.Alarm) error {
	deviceUUID, err := uuid.Parse(a.DeviceGUID)
	if err != nil {
		return fmt.Errorf("invalid device guid: %w", err)
	}

	var thresholdGUID pgtype.UUID
	if a.ThresholdGUID != nil {
		u, err := uuid.Parse(*a.ThresholdGUID)
		if err == nil {
			thresholdGUID = db.NewUUID(u)
		}
	}

	arg := db.CreateAlarmParams{
		DeviceGuid:     db.NewUUID(deviceUUID),
		ThresholdGuid:  thresholdGUID,
		ParameterName:  a.ParameterName,
		TriggeredValue: a.TriggeredValue,
		Severity:       db.AlarmSeverity(a.Severity),
		Status:         db.AlarmStatus(a.Status),
		TriggeredAt:    db.NewTimestamptz(a.TriggeredAt),
	}

	res, err := r.db.CreateAlarm(ctx, arg)
	if err != nil {
		return err
	}

	a.GUID = uuid.UUID(res.Guid.Bytes).String()
	a.CreatedAt = res.CreatedAt.Time
	a.UpdatedAt = res.UpdatedAt.Time
	return nil
}

func (r *alarmRepositoryImpl) GetByGUID(ctx context.Context, guidStr string) (*alarm.Alarm, error) {
	u, err := uuid.Parse(guidStr)
	if err != nil {
		return nil, alarm.ErrAlarmNotFound
	}

	res, err := r.db.GetAlarmByGUID(ctx, db.NewUUID(u))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, alarm.ErrAlarmNotFound
		}
		return nil, err
	}

	return db.AlarmFromDB(&res), nil
}

func (r *alarmRepositoryImpl) ListActiveByDevice(ctx context.Context, deviceGUID string) ([]*alarm.Alarm, error) {
	u, err := uuid.Parse(deviceGUID)
	if err != nil {
		return nil, nil
	}

	res, err := r.db.ListActiveAlarmsByDevice(ctx, db.NewUUID(u))
	if err != nil {
		return nil, err
	}

	var results []*alarm.Alarm
	for i := range res {
		results = append(results, db.AlarmFromDB(&res[i]))
	}
	return results, nil
}

func (r *alarmRepositoryImpl) ListHistory(ctx context.Context, params alarm.ListParams) (*alarm.ListResult, error) {
	var deviceUUID pgtype.UUID
	if params.DeviceGUID != nil && *params.DeviceGUID != "" && *params.DeviceGUID != "all" {
		u, err := uuid.Parse(*params.DeviceGUID)
		if err == nil {
			deviceUUID = db.NewUUID(u)
		}
	}

	var locationUUID pgtype.UUID
	if params.LocationGUID != nil && *params.LocationGUID != "" && *params.LocationGUID != "all" {
		u, err := uuid.Parse(*params.LocationGUID)
		if err == nil {
			locationUUID = db.NewUUID(u)
		}
	}

	var status db.NullAlarmStatus
	if params.Status != nil && *params.Status != "" && *params.Status != "all" {
		status = db.NullAlarmStatus{
			AlarmStatus: db.AlarmStatus(*params.Status),
			Valid:       true,
		}
	}

	offset := (params.Page - 1) * params.Limit
	arg := db.ListAlarmHistoryParams{
		Limit:        int32(params.Limit),
		Offset:       int32(offset),
		DeviceGuid:   deviceUUID,
		Status:       status,
		LocationGuid: locationUUID,
	}

	res, err := r.db.ListAlarmHistory(ctx, arg)
	if err != nil {
		return nil, err
	}

	total, err := r.db.CountAlarmHistory(ctx, db.CountAlarmHistoryParams{
		DeviceGuid:   deviceUUID,
		Status:       status,
		LocationGuid: locationUUID,
	})
	if err != nil {
		total = 0 // Fallback
	}

	var alarms []*alarm.Alarm
	for i := range res {
		a := &alarm.Alarm{
			GUID:           uuid.UUID(res[i].Guid.Bytes).String(),
			DeviceGUID:     uuid.UUID(res[i].DeviceGuid.Bytes).String(),
			DeviceSN:       res[i].DeviceSn,
			ParameterName:  res[i].ParameterName,
			TriggeredValue: res[i].TriggeredValue,
			Status:         string(res[i].Status),
			Severity:       string(res[i].Severity),
			TriggeredAt:    res[i].TriggeredAt.Time,
			CreatedAt:      res[i].CreatedAt.Time,
			UpdatedAt:      res[i].UpdatedAt.Time,
		}
		if res[i].DeviceAlias.Valid {
			a.DeviceAlias = res[i].DeviceAlias.String
		}
		a.LocationName = res[i].LocationName
		if res[i].ThresholdGuid.Valid {
			s := uuid.UUID(res[i].ThresholdGuid.Bytes).String()
			a.ThresholdGUID = &s
		}
		if res[i].ResolvedAt.Valid {
			a.ResolvedAt = &res[i].ResolvedAt.Time
		}
		if res[i].AcknowledgedAt.Valid {
			a.AcknowledgedAt = &res[i].AcknowledgedAt.Time
		}
		if res[i].AcknowledgedBy.Valid {
			s := uuid.UUID(res[i].AcknowledgedBy.Bytes).String()
			a.AcknowledgedBy = &s
		}
		alarms = append(alarms, a)
	}

	totalPages := 0
	if params.Limit > 0 {
		totalPages = int(total) / params.Limit
		if int(total)%params.Limit > 0 {
			totalPages++
		}
	}

	return &alarm.ListResult{
		Alarms:    alarms,
		Page:      params.Page,
		Limit:     params.Limit,
		Total:     total,
		TotalPage: totalPages,
	}, nil
}

func (r *alarmRepositoryImpl) GetStats(ctx context.Context, deviceGUID *string) (*alarm.Stats, error) {
	var dbID pgtype.UUID
	if deviceGUID != nil && *deviceGUID != "" && *deviceGUID != "all" {
		id, err := uuid.Parse(*deviceGUID)
		if err == nil {
			dbID = db.NewUUID(id)
		}
	}

	counts, err := r.db.CountAlarmsByStatus(ctx, dbID)
	if err != nil {
		return nil, err
	}

	stats := &alarm.Stats{}
	for _, c := range counts {
		stats.Total += c.Count
		switch c.Status {
		case db.AlarmStatusActive:
			stats.Active = c.Count
		case db.AlarmStatusAcknowledged:
			stats.Acknowledged = c.Count
		case db.AlarmStatusResolved:
			stats.Resolved = c.Count
		}
	}

	return stats, nil
}

func (r *alarmRepositoryImpl) UpdateStatus(ctx context.Context, guidStr string, status string, userIDStr *string) (*alarm.Alarm, error) {
	u, err := uuid.Parse(guidStr)
	if err != nil {
		return nil, alarm.ErrAlarmNotFound
	}

	var userUUID pgtype.UUID
	if userIDStr != nil {
		uu, err := uuid.Parse(*userIDStr)
		if err == nil {
			userUUID = db.NewUUID(uu)
		}
	}

	arg := db.UpdateAlarmStatusParams{
		Guid:           db.NewUUID(u),
		Status:         db.AlarmStatus(status),
		AcknowledgedBy: userUUID,
	}

	res, err := r.db.UpdateAlarmStatus(ctx, arg)
	if err != nil {
		return nil, err
	}

	return db.AlarmFromDB(&res), nil
}

func (r *alarmRepositoryImpl) GetActiveByDeviceParam(ctx context.Context, deviceGUID string, param string) (*alarm.Alarm, error) {
	u, err := uuid.Parse(deviceGUID)
	if err != nil {
		return nil, nil
	}

	res, err := r.db.GetActiveAlarmByDeviceParam(ctx, db.GetActiveAlarmByDeviceParamParams{
		DeviceGuid:    db.NewUUID(u),
		ParameterName: param,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return db.AlarmFromDB(&res), nil
}
