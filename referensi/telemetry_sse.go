package domain

import "time"

type TelemetryTagValue map[string]interface{}

type TelemetrySSE struct {
    SN        string                 `json:"sn"`
    Timestamp time.Time              `json:"timestamp"`
    Status    string                 `json:"status"`
    Telemetry map[string]interface{} `json:"telemetry"`
    Health    map[string]interface{} `json:"health"`
}

type TelemetrySSEResponse struct {
    Data *TelemetrySSE `json:"data"`
}
