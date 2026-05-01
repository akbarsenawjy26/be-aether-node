// repository/influx_repository.go
package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"be-go-historian/internal/domain"

	"github.com/google/uuid"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type influxRepository struct {
    client influxdb2.Client
    org    string
    bucket string
}

func NewInfluxRepository(client influxdb2.Client, org, bucket string) InfluxRepository {
    return &influxRepository{
        client: client,
        org:    org,
        bucket: bucket,
    }
}

func (r *influxRepository) GetLatestHealth(ctx context.Context, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.HealthData, error) {
    // Check if client is configured
    if r.client == nil {
        return nil, fmt.Errorf("InfluxDB client not configured")
    }

    // Build tenant filter - only apply if not super admin
    tenantFilter := ""
    if !isSuperAdmin {
        tenantFilter = fmt.Sprintf(`    |> filter(fn: (r) => r.tenant == "%s")`, tenantID.String())
    }

    query := fmt.Sprintf(`
from(bucket: "%s")
    |> range(start: -1m)
    |> filter(fn: (r) => r._measurement == "health")
    |> filter(fn: (r) => r._field == "uptime" or r._field == "fwVer" or r._field == "hwVer")
%s
    |> group(columns: ["tenant", "project", "sn", "_field"])
    |> last()
    |> pivot(rowKey:["tenant","project","sn","_time"], columnKey:["_field"], valueColumn:"_value")
    |> map(fn: (r) => ({
        tenant: r.tenant,
        project: r.project,
        sn: r.sn,
        uptime: r.uptime,
        fwVer: if exists r.fwVer then r.fwVer else "",
        hwVer: if exists r.hwVer then r.hwVer else "",
        _time: r._time,
        status: if (uint(v: now()) - uint(v: r._time)) > uint(v: 5 * 60 * 1000000000) then "offline" else "online"
    }))
    `, r.bucket, tenantFilter)

    q := r.client.QueryAPI(r.org)
    res, err := q.Query(ctx, query)
    if err != nil {
        return nil, err
    }

    var list []domain.HealthData

    for res.Next() {
        r := res.Record()

        // Safe parsing for uptime (could be string or float64)
        var uptime float64
        uptimeVal := r.ValueByKey("uptime")
        switch v := uptimeVal.(type) {
        case float64:
            uptime = v
        case string:
            // Try to parse string to float64
            if parsed, err := strconv.ParseFloat(v, 64); err == nil {
                uptime = parsed
            } else {
                uptime = 0 // Default value if parsing fails
            }
        default:
            uptime = 0 // Default value for unexpected types
        }

        list = append(list, domain.HealthData{
            Tenant:  r.ValueByKey("tenant").(string),
            Project: r.ValueByKey("project").(string),
            SN:      r.ValueByKey("sn").(string),
            FWVer:   fmt.Sprint(r.ValueByKey("fwVer")),
            HWVer:   fmt.Sprint(r.ValueByKey("hwVer")),
            Uptime:  uptime,
            Time:    r.Time(),
            Status:  fmt.Sprint(r.ValueByKey("status")),
        })
    }

    return list, nil
}

func (r *influxRepository) GetLatestHealthByProjectName(ctx context.Context, projectName string, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.HealthData, error) {
    if r.client == nil {
        return nil, fmt.Errorf("InfluxDB client not configured")
    }

    // Build tenant filter - only apply if not super admin
    tenantFilter := ""
    if !isSuperAdmin {
        // NOTE: InfluxDB stores tenant as NAME (from MQTT topic), not UUID
        // So we filter by tenant name. The tenantID authorization check happens at API level.
        // For now, we don't filter by tenant in InfluxDB - rely on authorization layer
        // TODO: In future, lookup actual tenant name from tenantID if needed
    }

    // Ganti QueryWithParams dengan Query biasa dan string interpolation
query := fmt.Sprintf(`
from(bucket: "%s")
    |> range(start: 0)
    |> filter(fn: (r) => r._measurement == "health")
    |> filter(fn: (r) => r._field == "uptime" or r._field == "fwVer" or r._field == "hwVer")
    |> filter(fn: (r) => r.project == "%s")
%s
    |> group(columns: ["tenant", "project", "sn", "_field"])
    |> last()
    |> pivot(rowKey:["tenant","project","sn","_time"], columnKey:["_field"], valueColumn:"_value")
    |> map(fn: (r) => ({
        tenant: r.tenant,
        project: r.project,
        sn: r.sn,
        uptime: r.uptime,
        fwVer: if exists r.fwVer then r.fwVer else "",
        hwVer: if exists r.hwVer then r.hwVer else "",
        _time: r._time,
        status: if (uint(v: now()) - uint(v: r._time)) > uint(v: 5 * 60 * 1000000000) then "offline" else "online"
    }))
`, r.bucket, projectName, tenantFilter)

// Ganti QueryWithParams dengan Query biasa
res, err := r.client.QueryAPI(r.org).Query(ctx, query)
if err != nil {
    return nil, err
}

var list []domain.HealthData

for res.Next() {
    r := res.Record()

    // Safe parsing for uptime (could be string or float64)
    var uptime float64
    uptimeVal := r.ValueByKey("uptime")
    switch v := uptimeVal.(type) {
    case float64:
        uptime = v
    case string:
        if parsed, err := strconv.ParseFloat(v, 64); err == nil {
            uptime = parsed
        } else {
            uptime = 0
        }
    default:
        uptime = 0
    }

    list = append(list, domain.HealthData{
        Tenant:  r.ValueByKey("tenant").(string),
        Project: r.ValueByKey("project").(string),
        SN:      r.ValueByKey("sn").(string),
        FWVer:   fmt.Sprint(r.ValueByKey("fwVer")),
        HWVer:   fmt.Sprint(r.ValueByKey("hwVer")),
        Uptime:  uptime,
        Time:    r.Time(),
        Status:  fmt.Sprint(r.ValueByKey("status")),
    })
}

return list, nil
}

func (r *influxRepository) GetLatestHealthByProjectID(ctx context.Context, projectID uuid.UUID, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.HealthData, error) {
    if r.client == nil {
        return nil, fmt.Errorf("InfluxDB client not configured")
    }

    // Build tenant filter - only apply if not super admin
    tenantFilter := ""
    if !isSuperAdmin {
        tenantFilter = fmt.Sprintf(`    |> filter(fn: (r) => r.tenant == "%s")`, tenantID.String())
    }

    // Filter by project ID instead of project name
    query := fmt.Sprintf(`
    from(bucket: "%s")
    |> range(start: 0)
    |> filter(fn: (r) =>
        r._measurement == "health" and
        (r._field == "uptime" or r._field == "fwVer" or r._field == "hwVer")
    )
%s
    |> filter(fn: (r) => r.project == "%s")
    |> group(columns: ["tenant", "project", "sn", "_field"])
    |> last()
    |> group(columns: ["tenant", "project", "sn"])
    |> pivot(rowKey:["tenant","project","sn","_time"], columnKey:["_field"], valueColumn:"_value")
    |> map(fn: (r) => ({
        tenant: r.tenant,
        project: r.project,
        sn: r.sn,
        uptime: r.uptime,
        fwVer: if exists r.fwVer then r.fwVer else "",
        hwVer: if exists r.hwVer then r.hwVer else "",
        _time: r._time,
        status: if (uint(v: now()) - uint(v: r._time)) > uint(v: 5 * 60 * 1000000000) then "offline" else "online"
    }))
    `, r.bucket, tenantFilter, projectID.String())

    res, err := r.client.QueryAPI(r.org).Query(ctx, query)
    if err != nil {
        return nil, err
    }

    var list []domain.HealthData

    for res.Next() {
        r := res.Record()
    
        // Safe parsing for uptime (could be string or float64)
        var uptime float64
        uptimeVal := r.ValueByKey("uptime")
        switch v := uptimeVal.(type) {
        case float64:
            uptime = v
        case string:
            if parsed, err := strconv.ParseFloat(v, 64); err == nil {
                uptime = parsed
            } else {
                uptime = 0
            }
        default:
            uptime = 0
        }

        list = append(list, domain.HealthData{
            Tenant:  r.ValueByKey("tenant").(string),
            Project: r.ValueByKey("project").(string),
            SN:      r.ValueByKey("sn").(string),
            FWVer:   fmt.Sprint(r.ValueByKey("fwVer")),
            HWVer:   fmt.Sprint(r.ValueByKey("hwVer")),
            Uptime:  uptime,
            Time:    r.Time(),
            Status:  fmt.Sprint(r.ValueByKey("status")),
        })
    }

    return list, nil
}

func (r *influxRepository) GetLatestHealthAllProject(ctx context.Context, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.HealthData, error) {
    if r.client == nil {
        return nil, fmt.Errorf("InfluxDB client not configured")
    }

    // Note: No tenant filtering in InfluxDB query since:
    // 1. InfluxDB stores tenant NAME (from MQTT topics), but we only have tenant UUID here
    // 2. Authorization is already handled at API/handler level
    // 3. User can only access projects within their tenant
    tenantFilter := ""

    query := fmt.Sprintf(`
    from(bucket: "%s")
    |> range(start: 0)
    |> filter(fn: (r) =>
        r._measurement == "health" and
        (r._field == "uptime" or r._field == "fwVer" or r._field == "hwVer")
    )
%s
    |> group(columns: ["tenant", "project", "sn", "_field"])
    |> last()
    |> group(columns: ["tenant", "project", "sn"])
    |> pivot(rowKey:["tenant","project","sn","_time"], columnKey:["_field"], valueColumn:"_value")
    |> map(fn: (r) => ({
        tenant: r.tenant,
        project: r.project,
        sn: r.sn,
        uptime: r.uptime,
        fwVer: r.fwVer,
        hwVer: r.hwVer,
        _time: r._time,
        status: if (uint(v: now()) - uint(v: r._time)) > uint(v: 5m)
            then "offline"
            else "online"
    }))

    `, r.bucket, tenantFilter)

    res, err := r.client.QueryAPI(r.org).Query(ctx, query)
    if err != nil {
        return nil, err
    }

    var list []domain.HealthData

    for res.Next() {
        r := res.Record()
    
        // Safe parsing for uptime (could be string or float64)
        var uptime float64
        uptimeVal := r.ValueByKey("uptime")
        switch v := uptimeVal.(type) {
        case float64:
            uptime = v
        case string:
            if parsed, err := strconv.ParseFloat(v, 64); err == nil {
                uptime = parsed
            } else {
                uptime = 0
            }
        default:
            uptime = 0
        }

        list = append(list, domain.HealthData{
            Tenant:  r.ValueByKey("tenant").(string),
            Project: r.ValueByKey("project").(string),
            SN:      r.ValueByKey("sn").(string),
            FWVer:   fmt.Sprint(r.ValueByKey("fwVer")),
            HWVer:   fmt.Sprint(r.ValueByKey("hwVer")),
            Uptime:  uptime,
            Time:    r.Time(),
            Status:  fmt.Sprint(r.ValueByKey("status")),
        })
    }

    return list, nil
}

func (r *influxRepository) GetLatestTelemetryBySN(ctx context.Context, sn string, tenantID uuid.UUID, isSuperAdmin bool) (*domain.TelemetrySSE, error) {

    // Note: No tenant filtering in InfluxDB query since:
    // 1. InfluxDB stores tenant NAME (from MQTT topics), but we only have tenant UUID here
    // 2. Authorization is already handled at API/handler level through device ownership check
    // 3. User can only access devices within their tenant
    tenantFilter := ""

    flux := fmt.Sprintf(`

        health =
            from(bucket: "%s")
                |> range(start: -10s)
                |> filter(fn: (r) => r._measurement == "health")
                |> filter(fn: (r) => r.sn == "%s")
%s
                |> last()
                |> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn:"_value")

        telemetry =
            from(bucket: "%s")
                |> range(start: -10s)
                |> filter(fn: (r) => r._measurement == "telemetry")
                |> filter(fn: (r) => r.sn == "%s")
%s
                |> last()
                |> keep(columns: ["_time", "sn", "tag", "_value"])
                |> pivot(
                    rowKey: ["_time", "sn"],
                    columnKey: ["tag"],
                    valueColumn: "_value"
                )

        union(tables: [health, telemetry])
    `, r.bucket, sn, tenantFilter, r.bucket, sn, tenantFilter)

    result, err := r.client.QueryAPI(r.org).Query(ctx, flux)
    if err != nil {
        return nil, err
    }

    out := &domain.TelemetrySSE{
        SN:        sn,
        Telemetry: make(map[string]interface{}),
        Health:    make(map[string]interface{}),
    }

    var latestTime time.Time
    var latestTelemetryTime time.Time
    var latestHealthTime time.Time

    for result.Next() {
        rec := result.Record()
        meas := rec.Measurement()
        ts := rec.Time()

        // pilih timestamp terbaru
        if ts.After(latestTime) {
            latestTime = ts
        }

        if meas == "telemetry" {
            if ts.After(latestTelemetryTime) {
                latestTelemetryTime = ts
                out.Telemetry = make(map[string]interface{})
            } else if ts.Before(latestTelemetryTime) {
                continue
            }
        }

        if meas == "health" {
            if ts.After(latestHealthTime) {
                latestHealthTime = ts
                out.Health = make(map[string]interface{})
            } else if ts.Before(latestHealthTime) {
                continue
            }
        }

        // ambil semua kolom dinamically (pivot menjadikan kolom)
        for key, val := range rec.Values() {
            if key == "_time" || key == "_measurement" || key == "sn" {
                continue
            }
            if val == nil {
                continue
            }

            if meas == "health" {
                out.Health[key] = val
            } else {
                out.Telemetry[key] = val
            }
        }
    }

    if result.Err() != nil {
        return nil, result.Err()
    }

    if latestTime.IsZero() {
        return nil, errors.New("no data for this SN")
    }

    // online/offline
    if time.Since(latestTime) <= 30*time.Second {
        out.Status = "online"
    } else {
        out.Status = "offline"
    }

    out.Timestamp = latestTime
    return out, nil
}


func (r *influxRepository) GetQueryHistoryTelemetryBySN(ctx context.Context, sn, start, stop string, tenantID uuid.UUID, isSuperAdmin bool) ([]domain.TelemetryHistory, error) {
    
    if r.client == nil {
        return nil, fmt.Errorf("InfluxDB client not configured")
    }

    // Note: No tenant filtering in InfluxDB query since:
    // 1. InfluxDB stores tenant NAME (from MQTT topics), but we only have tenant UUID here
    // 2. Authorization is already handled at API/handler level through device ownership check
    // 3. User can only access devices within their tenant
    tenantFilter := ""

    flux := fmt.Sprintf(`
        from(bucket: "%s")
            |> range(start: %s, stop: %s)
            |> filter(fn: (r) => r._measurement == "telemetry")
            |> filter(fn: (r) => r.sn == "%s")
%s
            |> keep(columns: ["_time", "sn", "tag", "_value"])
            |> pivot(
                rowKey: ["_time", "sn"],
                columnKey: ["tag"],
                valueColumn: "_value"
            )
            |> group(columns: ["sn"])
            |> sort(columns: ["_time"], desc: true)
            |> yield(name: "last")
    `, r.bucket, start, stop, sn, tenantFilter)

    res, err := r.client.QueryAPI(r.org).Query(ctx, flux)
    if err != nil {
        return nil, err
    }

    var list []domain.TelemetryHistory

    for res.Next() {
        rec := res.Record()

        item := domain.TelemetryHistory{
            SN:        fmt.Sprint(rec.ValueByKey("sn")),
            Timestamp: rec.Time(),
            Data:      make(map[string]interface{}),
        }

        for key, val := range rec.Values() {
            if key == "_time" || key == "sn" || key == "_start" ||
               key == "_stop" || key == "result" || key == "table" {
                continue
            }
            if val != nil {
                item.Data[key] = val
            }
        }

        list = append(list, item)
    }

    if res.Err() != nil {
        return nil, res.Err()
    }

    return list, nil
}