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
