-- name: CreateDevice :exec
INSERT INTO devices (guid, type, serial_number, alias, notes, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetDeviceByGUID :one
SELECT guid, type, serial_number, alias, notes, is_active, created_at, updated_at, deleted_at
FROM devices
WHERE guid = $1 AND deleted_at IS NULL;

-- name: GetDeviceBySerialNumber :one
SELECT guid, type, serial_number, alias, notes, is_active, created_at, updated_at, deleted_at
FROM devices
WHERE serial_number = $1 AND deleted_at IS NULL;

-- name: ListDevices :many
SELECT guid, type, serial_number, alias, notes, is_active, created_at, updated_at, deleted_at
FROM devices
WHERE deleted_at IS NULL
  AND (serial_number ILIKE $1 OR alias ILIKE $1 OR type ILIKE $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountDevices :one
SELECT COUNT(*) FROM devices WHERE deleted_at IS NULL;

-- name: UpdateDevice :exec
UPDATE devices
SET type = $2, serial_number = $3, alias = $4, notes = $5, is_active = $6, updated_at = $7
WHERE guid = $1 AND deleted_at IS NULL;

-- name: DeleteDevice :exec
UPDATE devices
SET deleted_at = $2, updated_at = $2
WHERE guid = $1 AND deleted_at IS NULL;

-- name: ExistsDeviceBySerialNumber :one
SELECT EXISTS(SELECT 1 FROM devices WHERE serial_number = $1 AND deleted_at IS NULL);
