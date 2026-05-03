# Server-Sent Events (SSE) — Reference Guide

## Apa itu SSE?

**Server-Sent Events (SSE)** adalah teknologi yang memungkinkan server mendorong data ke client melalui koneksi HTTP standar. Berbeda dengan WebSocket yang bersifat bidirectional (dua arah), SSE hanya satu arah: **server → client**.

```
┌──────────┐    HTTP Request (keep-alive)    ┌──────────┐
│  Client  │ ──────────────────────────────► │  Server  │
│          │ ◄────────────────────────────── │          │
│ (Browser)│    SSE Events (text/event-stream)│  (Go)    │
└──────────┘                                 └──────────┘
```

## Kapan Pakai SSE?

| use Case | SSE ✅ | WebSocket ❌ |
|----------|-------|-------------|
| Live data feed (IoT, dashboard) | ✅ | ✅ |
| Notifications | ✅ | ✅ |
| Chat real-time | ❌ | ✅ |
| Game multiplayer | ❌ | ✅ |
| Financial trading | ❌ | ✅ |

**IoT telemetry → SSE adalah pilihan yang tepat** karena device hanya mengirim data, tidak menerima dari client.

---

## Arsitektur SSE di Project Ini

```
┌─────────────────────────────────────────────────────────────┐
│                         Client                               │
│  (Browser / Postman / curl)                                │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP GET /stream
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              TelemetryHandler (Echo)                         │
│  • Set headers: text/event-stream, no-cache, keep-alive     │
│  • for-select loop: baca dari channel                       │
│  • fmt.Fprintf: "event: telemetry\ndata: {...}\n\n"          │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              TelemetryService (In-Memory Pub/Sub)           │
│  • subscribers map[deviceSN] → chan *Telemetry              │
│  • allDevices chan *Telemetry                               │
│  • startPublisher() goroutine — polling InfluxDB tiap 1s   │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              TelemetryRepository (InfluxDB)                  │
│  • GetAllLatest() — query latest point per device           │
│  • WriteTelemetry() — insert telemetry point               │
└─────────────────────────────────────────────────────────────┘
```

---

## 1. Handler Layer — `internal/handler/telemetry_handler.go`

### SSE Headers

```go
c.Response().Header().Set("Content-Type", "text/event-stream")
c.Response().Header().Set("Cache-Control", "no-cache")
c.Response().Header().Set("Connection", "keep-alive")
c.Response().Header().Set("Access-Control-Allow-Origin", "*")
c.Response().WriteHeader(http.StatusOK)
```

**Penjelasan tiap header:**

| Header | Nilai | Tujuan |
|--------|-------|--------|
| `Content-Type` | `text/event-stream` | Memberitahu client bahwa respons adalah SSE |
| `Cache-Control` | `no-cache` | Mencegah browser caching stream |
| `Connection` | `keep-alive` | Jaga koneksi tetap terbuka |
| `Access-Control-Allow-Origin` | `*` | Izinkan cross-origin request |

### SSE Format

```go
// Format SSE event:
fmt.Fprintf(c.Response(), "event: telemetry\ndata: %s\n\n", string(data))

// Komponen:
// - "event: <nama>"     → nama event (client bisa listen spesifik)
// - "data: <json>"      → payload data (bisa multi-line)
// - "\n\n"               → double newline = end of event (WAJIB)
```

**Contoh output ke client:**
```
event: telemetry
data: {"device_sn":"SN-001","temperature":25.5,"humidity":60.2,"timestamp":"2026-05-01T12:00:00Z"}

event: telemetry
data: {"device_sn":"SN-002","temperature":26.1,"humidity":58.8,"timestamp":"2026-05-01T12:00:01Z"}
```

### Keep-Alive (Heartbeat)

```go
ticker := time.NewTicker(30 * time.Second)
// ...
case <-ticker.C:
    fmt.Fprintf(c.Response(), ": keep-alive\n\n")
    flusher.Flush()
```

Setiap 30 detik, kirim comment `": "` supaya koneksi tidak timeout (proxy/load balancer biasanya timeout koneksi yang diam terlalu lama).

### Full Handler Pattern

```go
// StreamAllDevices handles GET /stream - SSE untuk semua device
func (h *TelemetryHandler) StreamAllDevices(c echo.Context) error {
    ctx := c.Request().Context()

    // 1. Set SSE headers
    c.Response().Header().Set("Content-Type", "text/event-stream")
    c.Response().Header().Set("Cache-Control", "no-cache")
    c.Response().Header().Set("Connection", "keep-alive")
    c.Response().Header().Set("Access-Control-Allow-Origin", "*")
    c.Response().WriteHeader(http.StatusOK)

    // 2. Context untuk cancel saat client disconnect
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // 3. Get channel dari service
    telemetryChan, errChan := h.svc.StreamAllDevices(ctx)

    // 4. Heartbeat ticker
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    // 5. for-select loop — event loop SSE
    for {
        select {
        case <-ctx.Done():
            // Client disconnect
            return nil

        case err := <-errChan:
            // Error dari service
            if err != nil {
                c.Logger().Error("SSE error: ", err)
            }
            return nil

        case t := <-telemetryChan:
            // Data dari channel → kirim ke client
            data, _ := json.Marshal(t)
            fmt.Fprintf(c.Response(), "event: telemetry\ndata: %s\n\n", string(data))
            if flusher, ok := c.Response().Writer.(http.Flusher); ok {
                flusher.Flush()
            }

        case <-ticker.C:
            // Keep-alive heartbeat
            fmt.Fprintf(c.Response(), ": keep-alive\n\n")
            if flusher, ok := c.Response().Writer.(http.Flusher); ok {
                flusher.Flush()
            }
        }
    }
}
```

**Kunci penting:**
- `defer cancel()` — kalau client disconnect, `ctx.Done()` ter-trigger, loop berhenti, tidak ada goroutine bocor
- `flusher.Flush()` — langsung kirim ke client, tidak tunggu buffer penuh
- `json.Marshal` di dalam loop — setiap data baru di marshal

---

## 2. Service Layer — `internal/service/telemetry/telemetry_service_impl.go`

### Pub/Sub Pattern dengan Channel

```go
type telemetryService struct {
    repo domainTelemetry.TelemetryRepository

    // SSE streaming
    subscribers map[string]chan *domainTelemetry.Telemetry  // per-device
    allDevices  chan *domainTelemetry.Telemetry             // broadcast
    mu          sync.RWMutex
}
```

**Kenapa pakai channel?**

- **Thread-safe** — tidak perlu mutex manual untuk mengirim data
- **Non-blocking** — select dengan default case, tidak pernah blok
- **Native Go** — tidak perlu library external

### Publisher Goroutine

```go
func (s *telemetryService) startPublisher() {
    ticker := time.NewTicker(1 * time.Second)  // poll tiap 1 detik
    defer ticker.Stop()

    for range ticker.C {
        data, err := s.repo.GetAllLatest(context.Background())
        if err != nil {
            continue
        }

        for _, t := range data {
            // Broadcast ke semua subscriber
            select {
            case s.allDevices <- t:
            default:
                // Buffer penuh, skip (jangan blok)
            }

            // Kirim ke subscriber spesifik device
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
```

**Tiap 1 detik:**
1. Query InfluxDB untuk dapat data latest semua device
2. Broadcast ke channel `allDevices` (untuk `/stream`)
3. Kirim ke subscriber spesifik (untuk `/stream/:device-sn`)

### StreamAllDevices — Broadcast Channel

```go
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
```

- Membuat channel baru untuk tiap request
- Goroutine membaca dari `allDevices` dan meneruskan ke channel client
- Kalau `ctx.Done()` (client disconnect), channel ditutup

### StreamDevice — Per-Device Subscription

```go
func (s *telemetryService) StreamDevice(ctx context.Context, deviceSN string) (<-chan *domainTelemetry.Telemetry, <-chan error) {
    telemetryChan := make(chan *domainTelemetry.Telemetry, 100)
    errChan := make(chan error, 1)

    // Register subscriber
    s.mu.Lock()
    s.subscribers[deviceSN] = telemetryChan
    s.mu.Unlock()

    // Cleanup saat ctx done
    go func() {
        <-ctx.Done()
        s.mu.Lock()
        delete(s.subscribers, deviceSN)
        s.mu.Unlock()
        close(telemetryChan)
    }()

    return telemetryChan, errChan
}
```

- Simpan channel ke `subscribers` map
- `publisher` goroutine otomatis kirim data ke subscriber terkait
- Saat ctx done → hapus dari map → tidak ada memory leak

### WriteTelemetry — Real-time Push

```go
func (s *telemetryService) WriteTelemetry(ctx context.Context, t *domainTelemetry.Telemetry) error {
    if err := s.repo.WriteTelemetry(ctx, t); err != nil {
        return err
    }

    // Push langsung tanpa tunggu polling 1 detik
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
```

**Kenapa ada goroutine di WriteTelemetry?**

- Data langsung dikirim ke SSE channel tanpa tunggu polling 1 detik
- User/device dapat real-time notification
- `select` dengan `default` = non-blocking, tidak mempengaruhi write performance

---

## 3. Repository Layer — InfluxDB

### Schema InfluxDB (Measurement: `environment`)

```
Tags:     device_sn, device_type, location_name
Fields:   temperature, humidity, aqi, pm25, pm10, co2, voc
Time:     timestamp
```

### GetAllLatest — Ambil Data Terbaru Tiap Device

*(Lihat file lengkap: `referensi/influx_repository_impl.go`)*

```go
// GetAllLatest returns the latest telemetry point for each device
func (r *influxRepository) GetAllLatest(ctx context.Context) ([]*domainTelemetry.Telemetry, error) {
    query := `
        SELECT LAST(temperature), LAST(humidity), LAST(aqi),
               LAST(pm25), LAST(pm10), LAST(co2), LAST(voc),
               LAST(device_type), LAST(location_name)
        FROM environment
        GROUP BY device_sn
    `

    // Parse result → []*domainTelemetry.Telemetry
    // ...
}
```

**Kenapa LAST() per group?**

- InfluxDB menyimpan semua data point
- `GROUP BY device_sn` → 1 row per device
- `LAST(*)` → nilai terbaru tiap field

---

## 4. Cara Testing SSE

### Postman

1. **GET** `http://localhost:8080/stream`
2. Klik **"Send"** — akan terbuka tab "Live Output"
3. Biarkan terbuka — akan terima SSE events
4. Kirim data: `POST /telemetry` → akan muncul di live output

### curl

```bash
curl -N -H "Accept: text/event-stream" http://localhost:8080/stream
```

Flags:
- `-N` — no buffer (langsung tampilkan output)
- `-H "Accept: text/event-stream"` — optional,明確

### Browser JavaScript

```javascript
// Koneksi SSE
const eventSource = new EventSource('http://localhost:8080/stream');

// Listen semua events
eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Device:', data.device_sn, 'Temp:', data.temperature);
};

// Listen event spesifik
eventSource.addEventListener('telemetry', function(event) {
    const data = JSON.parse(event.data);
    console.log('Specific handler:', data);
});

// Handle error
eventSource.onerror = function(err) {
    console.error('SSE Error:', err);
};

// Cleanup
// eventSource.close();
```

### Python

```python
import sseclient
import requests

response = requests.get('http://localhost:8080/stream', stream=True)
client = sseclient.SSEClient(response)

for event in client.events():
    print(event.data)
```

---

## 5. Pola Penting

### Non-Blocking Channel Send

```go
select {
case ch <- data:
default:
    // Buffer penuh, skip
}
```

**Jangan pernah** `ch <- data` tanpa select di dalam for-loop — akan deadlock kalau receiver lambat.

### Context untuk Cleanup

```go
ctx, cancel := context.WithCancel(ctx)
defer cancel()

// Di goroutine:
select {
case <-ctx.Done():
    // Cleanup di sini
    return
default:
}
```

**Harus ada context** di semua goroutine — kalau tidak, goroutine akan jalan terus meski client sudah disconnect (memory leak).

### Flusher Check

```go
if flusher, ok := c.Response().Writer.(http.Flusher); ok {
    flusher.Flush()
}
```

Tidak semua ResponseWriter punya interface `http.Flusher` (misalnya di test). Check type assertion dulu.

### SSE Event Format

```
event: <nama>\n
data: <json>\n
\n

```
- Satu event = 2 newline di akhir
- `event:` bisa omit → jadi anonymous event
- `data:` bisa multi-line kalau awali dengan indentation

---

## 6. Troubleshooting

### Client tidak dapat data

1. **Cek headers** — `Content-Type` harus `text/event-stream`
2. **Cek flusher** — `Flush()` harus dipanggil tiap event
3. **Cek keep-alive** — koneksi timeout karena proxy
4. **Cek browser console** — adakah error di Network tab?

### Goroutine leak

1. **Pastikan context di-cancel** saat client disconnect
2. **Gunakan `defer cancel()`** di awal handler
3. **Unregister subscriber** dari map saat ctx done

### Data tidak real-time

1. **Cek polling interval** — `startPublisher()` tiap 1 detik
2. **Cek buffer size** — channel buffer penuh bisa cause skip
3. **Cek non-blocking** — `default` case di select

### InfluxDB query slow

1. **Gunakan index** — `device_sn` harus tag, sudah di-index
2. **Gunakan LAST()** — tidak scan seluruh data
3. **Tambahkan retention policy** — hapus data lama otomatis

---

## 7. Ringkasan Pola SSE

```
Handler:
  headers SSE → Response.WriteHeader(200)
  for-select:
    case <-ctx.Done():    return
    case data := <-ch:    fmt.Fprintf("event: name\ndata: %s\n\n", json)
                          flusher.Flush()
    case <-ticker.C:      fmt.Fprintf(": keep-alive\n\n")
                          flusher.Flush()

Service:
  allDevices chan *Telemetry  (broadcast)
  subscribers map[sn] chan   (per-device)
  startPublisher() goroutine (poll repo tiap 1s)

Repository:
  GetAllLatest()           → query InfluxDB latest per device
  WriteTelemetry()         → insert ke InfluxDB
```

---

## 8. Pola Alternatif — Polling dengan json.Encoder

File referensi dari project lain (`influx_handler.go`, `influx_service_impl.go`) menggunakan pola berbeda:

### Handler — Pakai json.Encoder

```go
func (h *InfluxHandler) StreamTelemetryBySN(c echo.Context) error {
    // Set SSE headers
    res := c.Response()
    res.Header().Set("Content-Type", "text/event-stream")
    res.Header().Set("Cache-Control", "no-cache")
    res.Header().Set("Connection", "keep-alive")
    res.Header().Set("Access-Control-Allow-Origin", "*")

    // Cek flusher tersedia
    flusher, ok := res.Writer.(http.Flusher)
    if !ok {
        return echo.NewHTTPError(http.StatusInternalServerError, "Streaming not supported")
    }

    // Flush headers langsung
    flusher.Flush()

    ctx := c.Request().Context()
    dataCh := make(chan *domain.TelemetrySSE)

    go h.influxService.StreamTelemetryBySN(ctx, sn, tenantID, isSuperAdmin, dataCh)

    enc := json.NewEncoder(res)  // <--- json.Encoder

    for {
        select {
        case <-ctx.Done():
            return nil

        case data, ok := <-dataCh:
            // Channel closed → exit
            if !ok {
                return nil
            }
            if data == nil {
                continue  // Skip nil data
            }

            // Tulis data
            res.Write([]byte("data: "))
            enc.Encode(data)  // <--- Encode otomatis handle \n
            res.Write([]byte("\n\n"))
            flusher.Flush()
        }
    }
}
```

**Bedanya dengan Pattern 1:**

| Aspek | Pattern 1 (fmt.Fprintf) | Pattern 2 (json.Encoder) |
|-------|-------------------------|--------------------------|
| Encoding | Manual dengan `string(data)` | `enc.Encode()` langsung |
| Newline | Harus tambah `\n` manual | `Encode()` otomatis tambah `\n` |
| Error handling | `fmt.Fprintf` returns error | `enc.Encode()` returns error |
| Cleanup | `defer cancel()` | Check `ok` flag saat channel receive |

### Service — Polling dengan Retry

```go
func (s *influxService) StreamTelemetryBySN(ctx context.Context, sn string, tenantID uuid.UUID, isSuperAdmin bool, ch chan<- *domain.TelemetrySSE) {
    ticker := time.NewTicker(5 * time.Second)  // <--- 5 detik, bukan 1

    sendLatest := func() {
        data, err := s.influxRepo.GetLatestTelemetryBySN(ctx, sn, tenantID, isSuperAdmin)
        if err != nil {
            s.logger.Error().Err(err).Msg("Failed to fetch telemetry data")
            return  // JANGAN kirim nil, skip saja
        }
        ch <- data
    }

    // Kirim pertama kali langsung (tanpa tunggu ticker)
    sendLatest()

    for {
        select {
        case <-ctx.Done():
            ticker.Stop()
            close(ch)  // <--- Channel ditutup saat ctx done
            return

        case <-ticker.C:
            sendLatest()
        }
    }
}
```

**Bedanya:**

| Aspek | Pattern 1 (Publisher) | Pattern 2 (Polling) |
|-------|----------------------|---------------------|
| architecture | Pub/Sub central | Direct polling |
| Goroutine | 1 publisher semua client | 1 per-client |
| Interval | 1 detik | 5 detik |
| Channel | Shared `allDevices` | Per-client `dataCh` |
| Buffer | Non-blocking select | Blocking send (ticker controlled) |

### Error Handling — Empty Array vs Nil

```go
// Pattern 1: Skip on error
if err != nil {
    continue  // Skip, jangan kirim apa-apa
}

// Pattern 2: Kirim empty array (supaya client dapat response)
// Di influx_service_impl.go
if err != nil {
    s.logger.Error().Err(err).Msg("Failed to fetch health data from InfluxDB")
    ch <- []domain.HealthData{}  // Kosong, tapi tetap kirim
    continue
}
```

**Kapan pakai yang mana:**

- **Skip (`continue`)** — error transient, coba lagi next tick
- **Empty array** — client butuh konfirmasi bahwa request diproses

### Flusher Check & Immediate Flush

```go
// Check flusher tersedia
flusher, ok := c.Response().Writer.(http.Flusher)
if !ok {
    return echo.NewHTTPError(http.StatusInternalServerError, "Streaming not supported")
}

// Flush headers SEBELUM loop
flusher.Flush()
```

**Kenapa penting:**

- Beberapa proxy/CDN butuh headers flush duluan
- Tanpa flush, client bisa timeout sebelum data pertama

### Heartbeat Pattern

```go
// Pattern 1 — Comment-based
fmt.Fprintf(c.Response(), ": keep-alive\n\n")

// Pattern 2 — Ping event
c.Response().Write([]byte(": ping\n\n"))
c.Response().Flush()
```

Keduanya valid. Comment (`: ` prefix) adalah SSE spec standard untuk heartbeat.

---

## 9. Ringkasan Pola

```
PATTERN 1 — Pub/Sub (Concurrent)
─────────────────────────────────
Service:
  allDevices chan *Telemetry    (shared, broadcast)
  subscribers map[sn] chan      (per-device)
  startPublisher() goroutine    (1 detik, non-blocking)

  writeTelemetry():
    go func() {
      select { case s.allDevices <- t: default: }
      // dispatch ke subscribers
    }()

Handler:
  ctx, cancel := context.WithCancel(ctx)
  defer cancel()
  telemetryChan, _ := h.svc.StreamAllDevices(ctx)
  for-select { ... }


PATTERN 2 — Polling (Per-Client)
─────────────────────────────────
Service (per-client goroutine):
  ticker := time.NewTicker(5 * time.Second)
  sendLatest() // query repo
  for {
    select {
    case <-ctx.Done():
      ticker.Stop()
      close(ch)
      return
    case <-ticker.C:
      sendLatest()
    }
  }

Handler:
  flusher, ok := c.Response().Writer.(http.Flusher)
  flusher.Flush()  // flush headers dulu
  enc := json.NewEncoder(c.Response())
  for-select { ... }
```

---

## 10. Troubleshooting

### Client tidak dapat data

1. **Cek headers** — `Content-Type` harus `text/event-stream`
2. **Flush headers dulu** — `flusher.Flush()` sebelum loop
3. **Cek flusher** — `Flush()` harus dipanggil tiap event
4. **Cek proxy/load balancer** — timeout koneksi

### Goroutine leak

1. **Pattern 1** — `defer cancel()` di handler
2. **Pattern 2** — `ticker.Stop()` + `close(ch)` di `ctx.Done()`
3. **Unregister subscriber** dari map saat ctx done (Pattern 1)

### Data tidak real-time

1. **Cek polling interval** — 5 detik di Pattern 2
2. **Cek buffer size** — channel buffer penuh bisa cause skip
3. **Cek non-blocking** — `default` case di select

### InfluxDB query slow

1. **Gunakan index** — `device_sn` harus tag, sudah di-index
2. **Gunakan LAST()** — tidak scan seluruh data
3. **Tambahkan retention policy** — hapus data lama otomatis

---

## 11. Referensi

- **MDN SSE:** https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events
- **Echo Framework SSE:** https://echo.labstack.com/docs/streaming
- **InfluxDB:** https://docs.influxdata.com/influxdb/
- **Go Concurrency Patterns:** https://go.dev/blog/pipelines
- **be-go-historian SSE implementation:** `referensi/influx_handler.go`, `referensi/influx_service_impl.go`
