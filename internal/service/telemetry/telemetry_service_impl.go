package telemetry

import (
	"context"
	"sync"
	"time"

	domainTelemetry "aether-node/internal/domain/telemetry"
)

type telemetryService struct {
	repo domainTelemetry.TelemetryRepository

	// For SSE streaming
	subscribers map[string]chan *domainTelemetry.Telemetry
	allDevices  chan *domainTelemetry.Telemetry
	mu          sync.RWMutex
}

func NewTelemetryService(repo domainTelemetry.TelemetryRepository) domainTelemetry.TelemetryService {
	svc := &telemetryService{
		repo:        repo,
		subscribers: make(map[string]chan *domainTelemetry.Telemetry),
		allDevices:  make(chan *domainTelemetry.Telemetry, 100),
	}

	// Start the SSE publisher goroutine
	go svc.startPublisher()

	return svc
}

func (s *telemetryService) startPublisher() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data, err := s.repo.GetAllLatest(context.Background())
		if err != nil {
			continue
		}

		for _, t := range data {
		select {
			case s.allDevices <- t:
			default:
			}

			s.mu.RLock()
			if ch, ok := s.subscribers[t.DeviceSN]; ok {
				select {
				case ch <- t:
				default:
				}
			}
			s.mu.RUnlock()
		}
	}
}

func (s *telemetryService) WriteTelemetry(ctx context.Context, t *domainTelemetry.Telemetry) error {
	if err := s.repo.WriteTelemetry(ctx, t); err != nil {
		return err
	}

	go func() {
		select {
		case s.allDevices <- t:
		default:
		}

		s.mu.RLock()
		if ch, ok := s.subscribers[t.DeviceSN]; ok {
			select {
			case ch <- t:
			default:
			}
		}
		s.mu.RUnlock()
	}()

	return nil
}

func (s *telemetryService) StreamAllDevices(ctx context.Context) (<-chan *domainTelemetry.Telemetry, <-chan error) {
	telemetryChan := make(chan *domainTelemetry.Telemetry, 100)
	errChan := make(chan error, 1)

	go func() {
	for {
		select {
		case <-ctx.Done():
				close(telemetryChan)
			return
			case t := <-s.allDevices:
				telemetryChan <- t
			}
		}
	}()

	return telemetryChan, errChan
}

func (s *telemetryService) StreamDevice(ctx context.Context, deviceSN string) (<-chan *domainTelemetry.Telemetry, <-chan error) {
	telemetryChan := make(chan *domainTelemetry.Telemetry, 100)
	errChan := make(chan error, 1)

	s.mu.Lock()
	s.subscribers[deviceSN] = telemetryChan
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subscribers, deviceSN)
		s.mu.Unlock()
		close(telemetryChan)
	}()

	return telemetryChan, errChan
}

func (s *telemetryService) GetHistory(ctx context.Context, deviceSN string, query *domainTelemetry.TelemetryQuery) (*domainTelemetry.TelemetryListResult, error) {
	return s.repo.QueryHistory(ctx, deviceSN, query)
}

// ============================================================
// NEW: StreamAllDevicesWithHealth — merge health + telemetry, group by type
// Pattern dari DEVICE_STREAM_GUIDE.md Section 8
// ============================================================

func (s *telemetryService) StreamAllDevicesWithHealth(ctx context.Context, project string) (domainTelemetry.DevicePayload, error) {
	filter := domainTelemetry.DeviceFilter{Project: project}

	// Query health dan telemetry secara terpisah
	healthList, err := s.repo.GetLatestHealth(ctx, filter)
	if err != nil {
		return nil, err
	}

	telemetryMap, err := s.repo.GetLatestTelemetry(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Merge: health + telemetry → group by device type
	payload := make(domainTelemetry.DevicePayload)
	for _, h := range healthList {
		t := telemetryMap[h.DeviceSN]
		if t == nil {
			t = make(domainTelemetry.TelemetryData)
		}
		payload[h.Type] = append(payload[h.Type], domainTelemetry.DeviceEntry{
			Health:    h,
			Telemetry: t,
		})
	}

	return payload, nil
}

// ============================================================
// NEW: StreamDeviceWithHealth — merge health + telemetry untuk satu device
// Pattern dari DEVICE_STREAM_GUIDE.md Section 8
// ============================================================

func (s *telemetryService) StreamDeviceWithHealth(ctx context.Context, project, deviceSN string) (*domainTelemetry.DeviceEntry, error) {
	filter := domainTelemetry.DeviceFilter{Project: project, DeviceSN: deviceSN}

	// Query health dan telemetry secara terpisah
	healthList, err := s.repo.GetLatestHealth(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(healthList) == 0 {
		return nil, nil // device not found
	}

	telemetryMap, err := s.repo.GetLatestTelemetry(ctx, filter)
	if err != nil {
		return nil, err
	}

	t := telemetryMap[deviceSN]
	if t == nil {
		t = make(domainTelemetry.TelemetryData)
	}

	return &domainTelemetry.DeviceEntry{
		Health:    healthList[0],
		Telemetry: t,
	}, nil
}

// ============================================================
// NEW: GetTelemetryHistory — history dengan time range
// Pattern dari DEVICE_STREAM_GUIDE.md Section 8
// ============================================================

func (s *telemetryService) GetTelemetryHistory(ctx context.Context, filter domainTelemetry.HistoryFilter) ([]domainTelemetry.TelemetryRecord, error) {
	records, err := s.repo.GetTelemetryHistory(ctx, filter)
	if err != nil {
		return nil, err
	}
	if records == nil {
		return []domainTelemetry.TelemetryRecord{}, nil
	}
	return records, nil
}
