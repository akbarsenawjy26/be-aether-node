package alarm

import (
	"context"

	"aether-node/internal/domain/alarm"
)

type alarmServiceImpl struct {
	repo alarm.AlarmRepository
}

func NewAlarmService(repo alarm.AlarmRepository) alarm.AlarmService {
	return &alarmServiceImpl{repo: repo}
}

func (s *alarmServiceImpl) CreateAlarm(ctx context.Context, req *alarm.CreateAlarmRequest) (*alarm.Alarm, error) {
	a := &alarm.Alarm{
		DeviceGUID:     req.DeviceGUID,
		ThresholdGUID:  req.ThresholdGUID,
		ParameterName:  req.ParameterName,
		TriggeredValue: req.TriggeredValue,
		Severity:       req.Severity,
		Status:         req.Status,
		TriggeredAt:    req.TriggeredAt,
	}

	err := s.repo.Create(ctx, a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *alarmServiceImpl) GetAlarm(ctx context.Context, guid string) (*alarm.Alarm, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *alarmServiceImpl) ListActiveAlarms(ctx context.Context, deviceGUID string) ([]*alarm.Alarm, error) {
	return s.repo.ListActiveByDevice(ctx, deviceGUID)
}

func (s *alarmServiceImpl) ListAlarmHistory(ctx context.Context, params alarm.ListParams) (*alarm.ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	return s.repo.ListHistory(ctx, params)
}

func (s *alarmServiceImpl) AcknowledgeAlarm(ctx context.Context, guid string, userID string) (*alarm.Alarm, error) {
	return s.repo.UpdateStatus(ctx, guid, "acknowledged", &userID)
}

func (s *alarmServiceImpl) ResolveAlarm(ctx context.Context, guid string) (*alarm.Alarm, error) {
	return s.repo.UpdateStatus(ctx, guid, "resolved", nil)
}

func (s *alarmServiceImpl) GetAlarmStats(ctx context.Context, deviceGUID *string) (*alarm.Stats, error) {
	return s.repo.GetStats(ctx, deviceGUID)
}
