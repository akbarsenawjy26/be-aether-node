package telemetry

import (
	"context"
	"sync"
	"time"

	domainDevice "aether-node/internal/domain/device"
	domainTelemetry "aether-node/internal/domain/telemetry"
)

type telemetryService struct {
	repo       domainTelemetry.TelemetryRepository
	deviceRepo domainDevice.DeviceRepository

	// For SSE streaming
	subscribers map[string]chan *domainTelemetry.Telemetry
	allDevices  chan *domainTelemetry.Telemetry
	mu          sync.RWMutex

	// In-memory cache for device names (SN -> Alias)
	nameCache       map[string]string
	cacheMu         sync.RWMutex
	lastCacheUpdate time.Time
}

func NewTelemetryService(repo domainTelemetry.TelemetryRepository, deviceRepo domainDevice.DeviceRepository) domainTelemetry.TelemetryService {
	svc := &telemetryService{
		repo:        repo,
		deviceRepo:  deviceRepo,
		subscribers: make(map[string]chan *domainTelemetry.Telemetry),
		allDevices:  make(chan *domainTelemetry.Telemetry, 100),
		nameCache:   make(map[string]string),
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

func (s *telemetryService) getNameMap(ctx context.Context) map[string]string {
	s.cacheMu.RLock()
	// Refresh if cache is empty or older than 5 minutes
	if len(s.nameCache) > 0 && time.Since(s.lastCacheUpdate) < 5*time.Minute {
		defer s.cacheMu.RUnlock()
		// Return a copy to avoid race conditions if needed, but since we only replace the map, a reference is fine for RLock
		return s.nameCache
	}
	s.cacheMu.RUnlock()

	// Need update
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double check after lock
	if len(s.nameCache) > 0 && time.Since(s.lastCacheUpdate) < 5*time.Minute {
		return s.nameCache
	}

	devices, err := s.deviceRepo.List(ctx, domainDevice.ListParams{Limit: 1000})
	if err == nil && devices != nil {
		newMap := make(map[string]string)
		for _, d := range devices.Devices {
			newMap[d.SerialNumber] = d.Alias
		}
		s.nameCache = newMap
		s.lastCacheUpdate = time.Now()
	}
	return s.nameCache
}

func (s *telemetryService) InvalidateDeviceCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.nameCache = make(map[string]string)
	s.lastCacheUpdate = time.Time{} // Reset to zero time
}

func (s *telemetryService) StreamAllDevicesWithHealth(ctx context.Context, project string) (domainTelemetry.DevicePayload, error) {
	filter := domainTelemetry.DeviceFilter{Project: project}

	// 1. Get ALL devices from memory cache / Postgres (Master list)

	// 2. Query health and telemetry from InfluxDB
	healthList, err := s.repo.GetLatestHealthSSE(ctx, filter)
	if err != nil {
		return nil, err
	}

	telemetryMap, err := s.repo.GetLatestTelemetrySSE(ctx, filter)
	if err != nil {
		return nil, err
	}

	// 3. Merge data, using PostgreSQL as the base
	// (Previously we only iterated over healthList from Influx)
	
	// Pre-map health for easy lookup
	healthMap := make(map[string]domainTelemetry.HealthData)
	for _, h := range healthList {
		healthMap[h.DeviceSN] = h
	}

	payload := make(domainTelemetry.DevicePayload)
	
	// Get full device objects from DB to get their Type
	devices, err := s.deviceRepo.List(ctx, domainDevice.ListParams{Limit: 1000})
	if err != nil {
		return nil, err
	}

	for _, dev := range devices.Devices {
		// Filter by project if requested
		// Note: we might need project field in device model if we want strict filtering
		
		sn := dev.SerialNumber
		typeName := dev.Type
		if typeName == "" {
			typeName = "unknown"
		}

		// Get health from Influx or create default "offline" health
		health, ok := healthMap[sn]
		if !ok {
			health = domainTelemetry.HealthData{
				DeviceSN:   sn,
				DeviceName: dev.Alias,
				Type:       typeName,
				Status:     "offline",
			}
		} else {
			health.DeviceName = dev.Alias
		}

		// Get telemetry from Influx
		t := telemetryMap[sn]
		if t == nil {
			t = make(domainTelemetry.TelemetryData)
		}

		payload[typeName] = append(payload[typeName], domainTelemetry.DeviceEntry{
			Health:    health,
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

	// Query health dan telemetry secara terpisah (use SSE bypass methods to avoid circuit breaker)
	healthList, err := s.repo.GetLatestHealthSSE(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(healthList) == 0 {
		return nil, nil // device not found
	}

	telemetryMap, err := s.repo.GetLatestTelemetrySSE(ctx, filter)
	if err != nil {
		return nil, err
	}

	t := telemetryMap[deviceSN]
	if t == nil {
		t = make(domainTelemetry.TelemetryData)
	}

	health := healthList[0]
	// Fetch name from Postgres
	device, err := s.deviceRepo.GetBySerialNumber(ctx, deviceSN)
	if err == nil && device != nil {
		health.DeviceName = device.Alias
	}

	return &domainTelemetry.DeviceEntry{
		Health:    health,
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
