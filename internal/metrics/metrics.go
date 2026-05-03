package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// SSE metrics
	SSETotalConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sse_active_connections",
			Help: "Number of active SSE connections",
		},
	)

	SSETotalEventsSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sse_events_sent_total",
			Help: "Total number of SSE events sent",
		},
		[]string{"device_type"},
	)

	// InfluxDB metrics
	InfluxDBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "influxdb_query_duration_seconds",
			Help:    "InfluxDB query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"operation"},
	)

	InfluxDBQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "influxdb_query_errors_total",
			Help: "Total number of InfluxDB query errors",
		},
		[]string{"operation"},
	)

	// Auth metrics
	AuthLoginTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_total",
			Help: "Total number of login attempts",
		},
		[]string{"status"},
	)

	AuthTokenRefreshTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_token_refresh_total",
			Help: "Total number of token refresh attempts",
		},
		[]string{"status"},
	)

	// Device metrics
	DevicesOnline = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "devices_online_total",
			Help: "Number of online devices",
		},
	)

	DevicesByType = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "devices_by_type",
			Help: "Number of devices by type",
		},
		[]string{"type"},
	)
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path string, status int, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, statusLabel(status)).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordSSEConnection opened/closed
func RecordSSEConnection(opened bool) {
	if opened {
		SSETotalConnections.Inc()
	} else {
		SSETotalConnections.Dec()
	}
}

// RecordSSEEvent sent
func RecordSSEEvent(deviceType string) {
	SSETotalEventsSent.WithLabelValues(deviceType).Inc()
}

// RecordInfluxDBQuery records InfluxDB query metrics
func RecordInfluxDBQuery(operation string, duration float64, success bool) {
	InfluxDBQueryDuration.WithLabelValues(operation).Observe(duration)
	if !success {
		InfluxDBQueryErrors.WithLabelValues(operation).Inc()
	}
}

// RecordAuthLogin records login attempt
func RecordAuthLogin(success bool) {
	if success {
		AuthLoginTotal.WithLabelValues("success").Inc()
	} else {
		AuthLoginTotal.WithLabelValues("failure").Inc()
	}
}

// RecordDevicesCount updates device count metrics
func RecordDevicesCount(total int, byType map[string]int) {
	DevicesOnline.Set(float64(total))
	for dtype, count := range byType {
		DevicesByType.WithLabelValues(dtype).Set(float64(count))
	}
}

func statusLabel(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "2xx"
	case status >= 300 && status < 400:
		return "3xx"
	case status >= 400 && status < 500:
		return "4xx"
	case status >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
