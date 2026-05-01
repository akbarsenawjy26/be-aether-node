package telemetry

import (
	"context"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type telemetryService struct {
	repo         TelemetryRepository
	influxClient influxdb2.Client
	bucket       string
	org          string

	// For SSE streaming
	subscribers map[string]chan *Telemetry
	allDevices  chan *Telemetry
	mu          sync.RWMutex
}

func NewTelemetryService(
	repo TelemetryRepository,
	influxClient influxdb2.Client,
	org, bucket string,
) TelemetryService {
	svc := &telemetryService{
		repo:         repo,
		influxClient: influxClient,
		bucket:       bucket,
		org:          org,
		subscribers:  make(map[string]chan *Telemetry),
		allDevices:   make(chan *Telemetry, 100),
	}

	// Start the SSE publisher goroutine
	go svc.startPublisher()

	return svc
}

func (s *telemetryService) startPublisher() {
	// Query InfluxDB continuously for latest telemetry
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Get all latest telemetry data
		data, err := s.repo.GetAllLatest(context.Background())
		if err != nil {
			continue
		}

		for _, t := range data {
			// Broadcast to all-device stream
			select {
			case s.allDevices <- t:
			default:
			}

			// Broadcast to device-specific stream
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

func (s *telemetryService) WriteTelemetry(ctx context.Context, telemetry *Telemetry) error {
	// Write to InfluxDB
	if err := s.repo.WriteTelemetry(ctx, telemetry); err != nil {
		return err
	}

	// Broadcast to subscribers
	go func() {
		// To all devices
		select {
		case s.allDevices <- telemetry:
		default:
		}

		// To specific device
		s.mu.RLock()
		if ch, ok := s.subscribers[telemetry.DeviceSN]; ok {
			select {
			case ch <- telemetry:
			default:
			}
		}
		s.mu.RUnlock()
	}()

	return nil
}

func (s *telemetryService) StreamAllDevices(ctx context.Context) (<-chan *Telemetry, <-chan error) {
	telemetryChan := make(chan *Telemetry, 100)
	errChan := make(chan error, 1)

	// Subscribe to all devices stream
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

func (s *telemetryService) StreamDevice(ctx context.Context, deviceSN string) (<-chan *Telemetry, <-chan error) {
	telemetryChan := make(chan *Telemetry, 100)
	errChan := make(chan error, 1)

	// Create device-specific channel
	s.mu.Lock()
	s.subscribers[deviceSN] = telemetryChan
	s.mu.Unlock()

	// Cleanup when done
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subscribers, deviceSN)
		s.mu.Unlock()
		close(telemetryChan)
	}()

	return telemetryChan, errChan
}

func (s *telemetryService) GetHistory(ctx context.Context, deviceSN string, query *TelemetryQuery) (*TelemetryListResult, error) {
	return s.repo.QueryHistory(ctx, deviceSN, query)
}
