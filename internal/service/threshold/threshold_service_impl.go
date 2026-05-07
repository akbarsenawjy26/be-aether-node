package threshold

import (
	"context"
	"encoding/json"
	"time"

	"aether-node/internal/domain/threshold"

	"github.com/nats-io/nats.go"
)

type thresholdServiceImpl struct {
	repo threshold.ThresholdRepository
	nc   *nats.Conn
}

func NewThresholdService(repo threshold.ThresholdRepository, nc *nats.Conn) threshold.ThresholdService {
	return &thresholdServiceImpl{
		repo: repo,
		nc:   nc,
	}
}

func (s *thresholdServiceImpl) publishUpdate(action string, deviceGUID string) {
	if s.nc == nil {
		return
	}

	payload := map[string]interface{}{
		"action":      action,
		"device_guid": deviceGUID,
		"timestamp":   time.Now(),
	}
	data, _ := json.Marshal(payload)
	_ = s.nc.Publish("aether.threshold.updated", data)
}

func (s *thresholdServiceImpl) CreateThreshold(ctx context.Context, req *threshold.CreateThresholdRequest) (*threshold.Threshold, error) {
	t := &threshold.Threshold{
		DeviceGUID:    req.DeviceGUID,
		ParameterName: req.ParameterName,
		MinValue:      req.MinValue,
		MaxValue:      req.MaxValue,
		Severity:      req.Severity,
		IsActive:      req.IsActive,
	}

	err := s.repo.Create(ctx, t)
	if err != nil {
		return nil, err
	}

	s.publishUpdate("create", t.DeviceGUID)
	return t, nil
}

func (s *thresholdServiceImpl) GetThreshold(ctx context.Context, guid string) (*threshold.Threshold, error) {
	return s.repo.GetByGUID(ctx, guid)
}

func (s *thresholdServiceImpl) ListThresholdsByDevice(ctx context.Context, deviceGUID string) ([]*threshold.Threshold, error) {
	return s.repo.ListByDevice(ctx, deviceGUID)
}

func (s *thresholdServiceImpl) UpdateThreshold(ctx context.Context, guid string, req *threshold.UpdateThresholdRequest) (*threshold.Threshold, error) {
	t, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return nil, err
	}

	if req.ParameterName != nil {
		t.ParameterName = *req.ParameterName
	}
	if req.MinValue != nil {
		t.MinValue = req.MinValue
	}
	if req.MaxValue != nil {
		t.MaxValue = req.MaxValue
	}
	if req.Severity != nil {
		t.Severity = *req.Severity
	}
	if req.IsActive != nil {
		t.IsActive = *req.IsActive
	}

	err = s.repo.Update(ctx, t)
	if err != nil {
		return nil, err
	}

	s.publishUpdate("update", t.DeviceGUID)
	return t, nil
}

func (s *thresholdServiceImpl) DeleteThreshold(ctx context.Context, guid string) error {
	t, err := s.repo.GetByGUID(ctx, guid)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, guid)
	if err != nil {
		return err
	}

	s.publishUpdate("delete", t.DeviceGUID)
	return nil
}
