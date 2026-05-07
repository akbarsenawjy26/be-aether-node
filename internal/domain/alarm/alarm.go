package alarm

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAlarmNotFound = errors.New("alarm not found")
)

type Alarm struct {
	GUID           string     `json:"guid"`
	DeviceGUID     string     `json:"device_guid"`
	DeviceSN       string     `json:"device_sn,omitempty"`
	DeviceAlias    string     `json:"device_alias,omitempty"`
	LocationName   string     `json:"location_name,omitempty"`
	ThresholdGUID  *string    `json:"threshold_guid,omitempty"`
	ParameterName  string     `json:"parameter_name"`
	TriggeredValue float64    `json:"triggered_value"`
	Status         string     `json:"status"`
	Severity       string     `json:"severity"`
	TriggeredAt    time.Time  `json:"triggered_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	AcknowledgedBy *string    `json:"acknowledged_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type ListParams struct {
	DeviceGUID   *string
	LocationGUID *string
	Status       *string
	Limit        int
	Page         int
}

type ListResult struct {
	Alarms    []*Alarm
	Total     int64
	Page      int
	Limit     int
	TotalPage int
}

type CreateAlarmRequest struct {
	DeviceGUID     string    `json:"device_guid"`
	ThresholdGUID  *string   `json:"threshold_guid"`
	ParameterName  string    `json:"parameter_name"`
	TriggeredValue float64   `json:"triggered_value"`
	Severity       string    `json:"severity"`
	Status         string    `json:"status"`
	TriggeredAt    time.Time `json:"triggered_at"`
}

type Stats struct {
	Total        int64 `json:"total"`
	Active       int64 `json:"active"`
	Acknowledged int64 `json:"acknowledged"`
	Resolved     int64 `json:"resolved"`
}

type AlarmRepository interface {
	Create(ctx context.Context, a *Alarm) error
	GetByGUID(ctx context.Context, guid string) (*Alarm, error)
	ListActiveByDevice(ctx context.Context, deviceGUID string) ([]*Alarm, error)
	ListHistory(ctx context.Context, params ListParams) (*ListResult, error)
	UpdateStatus(ctx context.Context, guid string, status string, userID *string) (*Alarm, error)
	GetActiveByDeviceParam(ctx context.Context, deviceGUID string, param string) (*Alarm, error)
	GetStats(ctx context.Context, deviceGUID *string) (*Stats, error)
}

type AlarmService interface {
	CreateAlarm(ctx context.Context, req *CreateAlarmRequest) (*Alarm, error)
	GetAlarm(ctx context.Context, guid string) (*Alarm, error)
	ListActiveAlarms(ctx context.Context, deviceGUID string) ([]*Alarm, error)
	ListAlarmHistory(ctx context.Context, params ListParams) (*ListResult, error)
	AcknowledgeAlarm(ctx context.Context, guid string, userID string) (*Alarm, error)
	ResolveAlarm(ctx context.Context, guid string) (*Alarm, error)
	GetAlarmStats(ctx context.Context, deviceGUID *string) (*Stats, error)
}
