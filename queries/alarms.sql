-- name: CreateAlarm :one
INSERT INTO alarms (
    device_guid, threshold_guid, parameter_name, triggered_value, severity, status, triggered_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetAlarmByGUID :one
SELECT * FROM alarms
WHERE guid = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListActiveAlarmsByDevice :many
SELECT * FROM alarms
WHERE device_guid = $1 AND status = 'active' AND deleted_at IS NULL
ORDER BY triggered_at DESC;

-- name: ListAlarmHistory :many
SELECT 
    a.*, 
    d.serial_number as device_sn, 
    d.alias as device_alias,
    COALESCE((SELECT l.name FROM installation_points ip 
      JOIN locations l ON ip.location_guid = l.guid 
      WHERE ip.device_guid = a.device_guid AND ip.deleted_at IS NULL 
      ORDER BY ip.created_at DESC LIMIT 1), '')::text as location_name
FROM alarms a
JOIN devices d ON a.device_guid = d.guid
WHERE (a.device_guid = sqlc.narg('device_guid') OR sqlc.narg('device_guid') IS NULL)
  AND (a.status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
  AND (sqlc.narg('location_guid')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM installation_points ip 
      WHERE ip.device_guid = a.device_guid 
        AND ip.location_guid = sqlc.narg('location_guid') 
        AND ip.deleted_at IS NULL
  ))
  AND a.deleted_at IS NULL
ORDER BY a.triggered_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAlarmHistory :one
SELECT count(*) 
FROM alarms a
WHERE (a.device_guid = sqlc.narg('device_guid') OR sqlc.narg('device_guid') IS NULL)
  AND (a.status = sqlc.narg('status') OR sqlc.narg('status') IS NULL)
  AND (sqlc.narg('location_guid')::uuid IS NULL OR EXISTS (
      SELECT 1 FROM installation_points ip 
      WHERE ip.device_guid = a.device_guid 
        AND ip.location_guid = sqlc.narg('location_guid') 
        AND ip.deleted_at IS NULL
  ))
  AND a.deleted_at IS NULL;

-- name: CountAlarmsByStatus :many
SELECT status, count(*) as count
FROM alarms
WHERE (device_guid = $1 OR $1 IS NULL)
  AND deleted_at IS NULL
GROUP BY status;

-- name: UpdateAlarmStatus :one
UPDATE alarms
SET 
    status = sqlc.arg('status')::alarm_status,
    acknowledged_at = CASE WHEN sqlc.arg('status')::text = 'acknowledged' THEN NOW() ELSE acknowledged_at END,
    acknowledged_by = CASE WHEN sqlc.arg('status')::text = 'acknowledged' THEN sqlc.arg('acknowledged_by') ELSE acknowledged_by END,
    resolved_at = CASE WHEN sqlc.arg('status')::text = 'resolved' THEN NOW() ELSE resolved_at END,
    updated_at = NOW()
WHERE guid = sqlc.arg('guid') AND deleted_at IS NULL
RETURNING *;

-- name: GetActiveAlarmByDeviceParam :one
SELECT * FROM alarms
WHERE device_guid = $1 AND parameter_name = $2 AND status = 'active' AND deleted_at IS NULL
LIMIT 1;
