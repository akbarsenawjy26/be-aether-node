// domain/health_sse.go
package domain

import "time"

type HealthData struct {
    Tenant  string    `json:"tenant"`
    Project string    `json:"project"`
    SN      string    `json:"sn"`
    FWVer   string    `json:"fwVer"`
    HWVer   string    `json:"hwVer"`
    Uptime  float64   `json:"uptime"`
    Time    time.Time `json:"time"`
    Status  string    `json:"status"`
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
    ID    string `json:"id"`
    Event string `json:"event"`
    Data  string `json:"data"`
}
