// services/influx_service.go
package services

import (
	"be-go-historian/internal/domain"
	"context"

	"github.com/google/uuid"
)

type InfluxService interface {
    StreamHealth(ctx context.Context, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData)
	StreamHealthByProjectName(ctx context.Context, projectName string, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData)
	StreamHealthByProjectID(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData)
	StreamHealthAllProject(ctx context.Context, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- []domain.HealthData)
	StreamTelemetryBySN(ctx context.Context, sn string, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- *domain.TelemetrySSE)
	GetQueryHistoryTelemetryBySN(ctx context.Context, sn, start, stop string, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.TelemetryHistory, error)
}
