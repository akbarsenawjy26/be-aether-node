package services

import (
	"context"
	"time"

	"be-go-historian/internal/domain"
	"be-go-historian/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type influxService struct {
	influxRepo repository.InfluxRepository
	logger     zerolog.Logger
}

func NewInfluxService(repo repository.InfluxRepository, logger zerolog.Logger) InfluxService {
    return &influxService{
        influxRepo: repo,
        logger:     logger,
    }
}

func (s *influxService) StreamHealth(ctx context.Context, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData) {
    data, err := s.influxRepo.GetLatestHealth(ctx, tenantID, isSuperAdmin)
            if err != nil {
                s.logger.Error().Err(err).Msg("Failed to fetch health data from InfluxDB")
                // Send empty array instead of nil to prevent null responses
                ch <- []domain.HealthData{}
            } else {
            s.logger.Debug().Int("records", len(data)).Msg("Fetched health data from InfluxDB")
            ch <- data
        }
    
    ticker := time.NewTicker(5 * time.Second) // refresh setiap 5 detik
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            close(ch)
            return
        case <-ticker.C:
            data, err := s.influxRepo.GetLatestHealth(ctx, tenantID, isSuperAdmin)
            if err != nil {
                s.logger.Error().Err(err).Msg("Failed to fetch health data from InfluxDB")
                // Send empty array instead of nil to prevent null responses
                ch <- []domain.HealthData{}
                continue
            }
            s.logger.Debug().Int("records", len(data)).Msg("Fetched health data from InfluxDB")
            ch <- data
        }
    }
}

func (s *influxService) StreamHealthByProjectName(ctx context.Context, projectName string, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData) {
    data, err := s.influxRepo.GetLatestHealthByProjectName(ctx, projectName, tenantID, isSuperAdmin)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to fetch initial health data")
        ch <- []domain.HealthData{}
    } else {
        s.logger.Debug().Int("records", len(data)).Msg("Fetched initial health data")
        ch <- data
    }
        
    ticker := time.NewTicker(5 * time.Second) // refresh setiap 5 detik
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            close(ch)
            return
        case <-ticker.C:
            data, err := s.influxRepo.GetLatestHealthByProjectName(ctx, projectName, tenantID, isSuperAdmin)
            if err != nil {
                s.logger.Error().Err(err).Msg("Failed to fetch health data from InfluxDB")
                // Send empty array instead of nil to prevent null responses
                ch <- []domain.HealthData{}
                continue
            }
            s.logger.Debug().Int("records", len(data)).Msg("Fetched health data from InfluxDB")
            ch <- data
        }
    }
}

func (s *influxService) StreamHealthByProjectID(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData) {
    data, err := s.influxRepo.GetLatestHealthByProjectID(ctx, projectID, tenantID, isSuperAdmin)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to fetch initial health data")
        ch <- []domain.HealthData{}
    } else {
        s.logger.Debug().Int("records", len(data)).Msg("Fetched initial health data")
        ch <- data
    }
        
    ticker := time.NewTicker(5 * time.Second) // refresh setiap 5 detik
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            close(ch)
            return
        case <-ticker.C:
            data, err := s.influxRepo.GetLatestHealthByProjectID(ctx, projectID, tenantID, isSuperAdmin)
            if err != nil {
                s.logger.Error().Err(err).Msg("Failed to fetch health data from InfluxDB")
                // Send empty array instead of nil to prevent null responses
                ch <- []domain.HealthData{}
                continue
            }
            s.logger.Debug().Int("records", len(data)).Msg("Fetched health data from InfluxDB")
            ch <- data
        }
    }
}

func (s *influxService) StreamHealthAllProject(ctx context.Context, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData) {
    // Query pertama langsung tanpa menunggu ticker
    data, err := s.influxRepo.GetLatestHealthAllProject(ctx, tenantID, isSuperAdmin)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to fetch initial health data")
        ch <- []domain.HealthData{}
    } else {
        s.logger.Debug().Int("records", len(data)).Msg("Fetched initial health data")
        ch <- data
    }

    // Baru mulai ticker untuk update selanjutnya
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            close(ch)
            return
        case <-ticker.C:
            data, err := s.influxRepo.GetLatestHealthAllProject(ctx, tenantID, isSuperAdmin)
            if err != nil {
                s.logger.Error().Err(err).Msg("Failed to fetch health data from InfluxDB")
                ch <- []domain.HealthData{}
                continue
            }
            s.logger.Debug().Int("records", len(data)).Msg("Fetched health data from InfluxDB")
            ch <- data
        }
    }
}

func (s *influxService) StreamTelemetryBySN(ctx context.Context, sn string, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- *domain.TelemetrySSE) {
    ticker := time.NewTicker(5 * time.Second)

    // fungsi fetch data
    sendLatest := func() {
        data, err := s.influxRepo.GetLatestTelemetryBySN(ctx, sn, tenantID, isSuperAdmin)
        if err != nil {
            s.logger.Error().Err(err).Msg("Failed to fetch telemetry data from InfluxDB")
            return // JANGAN kirim nil
        }
        ch <- data
    }

    // kirim pertama kali
    sendLatest()

    for {
        select {
        case <-ctx.Done():
            ticker.Stop()
            close(ch) // channel ditutup hanya saat loop selesai
            return

        case <-ticker.C:
            sendLatest()
        }
    }
}

func (s *influxService) GetQueryHistoryTelemetryBySN(ctx context.Context, sn, start, stop string, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.TelemetryHistory, error) {
    data, err := s.influxRepo.GetQueryHistoryTelemetryBySN(ctx, sn, start, stop, tenantID, isSuperAdmin)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to fetch telemetry history data from InfluxDB")
        return nil, err
    }
    s.logger.Debug().Int("records", len(data)).Msg("Fetched telemetry history data from InfluxDB")
    return data, nil
}