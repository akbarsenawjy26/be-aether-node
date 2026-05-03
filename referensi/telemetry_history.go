package domain

import "time"

type TelemetryHistory struct {
    SN string `json:"sn"`
    Timestamp time.Time `json:"timestamp"`
    Data map[string]interface{} `json:"data"`
}

type TelemetryHistoryResponse struct {
    Data *TelemetryHistory `json:"data"`
}