# Grafana Dashboard Guide - Aether Node

## Metrics Endpoint
```
GET http://localhost:8085/metrics
```

---

## Available Metrics

### 1. HTTP Metrics

#### `http_requests_total` (Counter)
Total HTTP requests.
- **Labels:** `method`, `path`, `status`
- **Query:** `rate(http_requests_total[5m])`

#### `http_request_duration_seconds` (Histogram)
HTTP request latency distribution.
- **Labels:** `method`, `path`
- **Query:** `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`

---

### 2. SSE (Server-Sent Events) Metrics

#### `sse_active_connections` (Gauge)
Current number of active SSE connections.
- **Query:** `sse_active_connections`

#### `sse_events_sent_total` (Counter)
Total SSE events sent to clients.
- **Labels:** `device_type`
- **Query:** `rate(sse_events_sent_total[5m])`

---

### 3. InfluxDB Metrics

#### `influxdb_query_duration_seconds` (Histogram)
InfluxDB query latency.
- **Labels:** `operation`
  - `get_latest_health`
  - `get_latest_telemetry`
  - `get_telemetry_history`
- **Query:** `histogram_quantile(0.95, rate(influxdb_query_duration_seconds_bucket[5m]))`

#### `influxdb_query_errors_total` (Counter)
Total InfluxDB query errors.
- **Labels:** `operation`
- **Query:** `rate(influxdb_query_errors_total[5m])`

---

### 4. Auth Metrics

#### `auth_login_total` (Counter)
Login attempts.
- **Labels:** `status` — `success`, `failure`
- **Query:** `rate(auth_login_total[5m])`

#### `auth_token_refresh_total` (Counter)
Token refresh attempts.
- **Labels:** `status` — `success`, `failure`

---

### 5. Device Metrics

#### `devices_online_total` (Gauge)
Number of online devices.
- **Query:** `devices_online_total`

#### `devices_by_type` (Gauge)
Number of devices by type.
- **Labels:** `type` — `aqi`, `bacis`, etc.
- **Query:** `devices_by_type`

---

## Suggested Grafana Dashboard Panels

### Row 1: HTTP Overview

| Panel Name | Visualization | Query |
|------------|---------------|-------|
| **Request Rate** | Time series | `sum(rate(http_requests_total[5m])) by (status)` |
| **Request Latency (p95)** | Time series | `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))` |
| **Error Rate (4xx+5xx)** | Time series | `sum(rate(http_requests_total{status=~"4..|5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100` |
| **Requests by Path** | Pie chart | `sum(rate(http_requests_total[5m])) by (path)` |

### Row 2: SSE / Telemetry

| Panel Name | Visualization | Query |
|------------|---------------|-------|
| **Active SSE Connections** | Stat | `sse_active_connections` |
| **SSE Events Rate** | Time series | `sum(rate(sse_events_sent_total[5m])) by (device_type)` |
| **SSE Events by Type** | Time series | `rate(sse_events_sent_total[5m])` |
| **Avg SSE Latency (p50)** | Time series | `histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{path=~"/telemetry/stream.*"}[5m]))` |

### Row 3: InfluxDB Performance

| Panel Name | Visualization | Query |
|------------|---------------|-------|
| **InfluxDB Query Latency (p95)** | Time series | `histogram_quantile(0.95, rate(influxdb_query_duration_seconds_bucket[5m])) by (operation)` |
| **InfluxDB Query Rate** | Time series | `sum(rate(influxdb_query_duration_seconds_count[5m])) by (operation)` |
| **InfluxDB Error Rate** | Time series | `sum(rate(influxdb_query_errors_total[5m])) by (operation)` |
| **Slow Queries (>100ms)** | Time series | `sum(rate(influxdb_query_duration_seconds_bucket{le="0.1"}[5m])) by (operation) / sum(rate(influxdb_query_duration_seconds_count[5m])) by (operation)` |

### Row 4: Auth

| Panel Name | Visualization | Query |
|------------|---------------|-------|
| **Login Success Rate** | Time series | `sum(rate(auth_login_total{status="success"}[5m])) / sum(rate(auth_login_total[5m])) * 100` |
| **Login Attempts** | Time series | `sum(rate(auth_login_total[5m])) by (status)` |
| **Failed Logins** | Stat | `sum(increase(auth_login_total{status="failure"}[1h]))` |

### Row 5: Devices

| Panel Name | Visualization | Query |
|------------|---------------|-------|
| **Online Devices** | Stat | `devices_online_total` |
| **Devices by Type** | Pie chart | `devices_by_type` |
| **Devices Trend** | Time series | `devices_online_total` |

---

## Grafana Dashboard JSON

```json
{
  "annotations": {
    "list": []
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "title": "Request Rate",
      "type": "timeseries",
      "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0},
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[5m])) by (status)",
          "legendFormat": "{{status}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "reqps",
          "custom": {"drawStyle": "Line"}
        }
      }
    },
    {
      "title": "Request Latency p95",
      "type": "timeseries",
      "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0},
      "targets": [
        {
          "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
          "legendFormat": "{{method}} {{path}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "s",
          "custom": {"drawStyle": "Line"}
        }
      }
    },
    {
      "title": "Active SSE Connections",
      "type": "stat",
      "gridPos": {"h": 4, "w": 4, "x": 0, "y": 8},
      "targets": [
        {
          "expr": "sse_active_connections"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "none",
          "thresholds": {
            "steps": [
              {"value": 0, "color": "green"},
              {"value": 10, "color": "yellow"},
              {"value": 50, "color": "red"}
            ]
          }
        }
      }
    },
    {
      "title": "SSE Events Rate",
      "type": "timeseries",
      "gridPos": {"h": 4, "w": 8, "x": 4, "y": 8},
      "targets": [
        {
          "expr": "sum(rate(sse_events_sent_total[5m])) by (device_type)",
          "legendFormat": "{{device_type}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "eps",
          "custom": {"drawStyle": "Line"}
        }
      }
    },
    {
      "title": "InfluxDB Query Latency p95",
      "type": "timeseries",
      "gridPos": {"h": 4, "w": 12, "x": 12, "y": 8},
      "targets": [
        {
          "expr": "histogram_quantile(0.95, rate(influxdb_query_duration_seconds_bucket[5m])) by (operation)",
          "legendFormat": "{{operation}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "s",
          "custom": {"drawStyle": "Line"}
        }
      }
    },
    {
      "title": "Online Devices",
      "type": "stat",
      "gridPos": {"h": 4, "w": 4, "x": 0, "y": 12},
      "targets": [
        {
          "expr": "devices_online_total"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "none"
        }
      }
    },
    {
      "title": "Devices by Type",
      "type": "piechart",
      "gridPos": {"h": 4, "w": 8, "x": 4, "y": 12},
      "targets": [
        {
          "expr": "devices_by_type"
        }
      ]
    },
    {
      "title": "Login Success Rate",
      "type": "gauge",
      "gridPos": {"h": 4, "w": 6, "x": 12, "y": 12},
      "targets": [
        {
          "expr": "sum(rate(auth_login_total{status=\"success\"}[5m])) / sum(rate(auth_login_total[5m])) * 100"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "thresholds": {
            "steps": [
              {"value": 0, "color": "red"},
              {"value": 80, "color": "yellow"},
              {"value": 95, "color": "green"}
            ]
          }
        }
      }
    },
    {
      "title": "Failed Logins (1h)",
      "type": "stat",
      "gridPos": {"h": 4, "w": 6, "x": 18, "y": 12},
      "targets": [
        {
          "expr": "sum(increase(auth_login_total{status=\"failure\"}[1h]))"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "none",
          "thresholds": {
            "steps": [
              {"value": 0, "color": "green"},
              {"value": 5, "color": "yellow"},
              {"value": 10, "color": "red"}
            ]
          }
        }
      }
    },
    {
      "title": "InfluxDB Error Rate",
      "type": "timeseries",
      "gridPos": {"h": 8, "w": 24, "x": 0, "y": 16},
      "targets": [
        {
          "expr": "sum(rate(influxdb_query_errors_total[5m])) by (operation)",
          "legendFormat": "{{operation}}"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "eps",
          "custom": {"drawStyle": "Line"}
        }
      }
    }
  ],
  "refresh": "5s",
  "schemaVersion": 38,
  "tags": ["aether-node", "telemetry"],
  "templating": {"list": []},
  "time": {"from": "now-1h", "to": "now"},
  "timepicker": {},
  "timezone": "browser",
  "title": "Aether Node Dashboard",
  "uid": "aether-node",
  "version": 1
}
```

---

## Prometheus Datasource Setup

1. Go to **Configuration → Data Sources**
2. Add **Prometheus**
3. URL: `http://localhost:9090` (or your Prometheus server)
4. Click **Save & Test**

---

## Alert Rules (Optional)

### High Error Rate
```yaml
alert: HighHTTPErrorRate
expr: sum(rate(http_requests_total{status=~"5.."}[5m])) > 0.05
for: 5m
labels:
  severity: critical
annotations:
  summary: "High HTTP 5xx error rate"
```

### High SSE Connection Count
```yaml
alert: HighSSEConnections
expr: sse_active_connections > 100
for: 5m
labels:
  severity: warning
annotations:
  summary: "High number of SSE connections"
```

### Slow InfluxDB Queries
```yaml
alert: SlowInfluxDBQueries
expr: histogram_quantile(0.95, rate(influxdb_query_duration_seconds_bucket[5m])) > 0.5
for: 5m
labels:
  severity: warning
annotations:
  summary: "InfluxDB p95 query latency > 500ms"
```

### Failed Login Spike
```yaml
alert: FailedLoginSpike
expr: sum(rate(auth_login_total{status="failure"}[5m])) > 10
for: 5m
labels:
  severity: warning
annotations:
  summary: "High number of failed login attempts"
```
