-- name: CreateInstallationPoint :exec
INSERT INTO installation_points (guid, name, device_guid, location_guid, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetInstallationPointByGUID :one
SELECT guid, name, device_guid, location_guid, notes, created_at, updated_at, deleted_at
FROM installation_points
WHERE guid = $1 AND deleted_at IS NULL;

-- name: ListInstallationPoints :many
SELECT 
    ip.guid, ip.name, ip.device_guid, ip.location_guid, ip.notes, ip.created_at, ip.updated_at, ip.deleted_at,
    d.serial_number, d.alias, l.name as location_name
FROM installation_points ip
LEFT JOIN devices d ON ip.device_guid = d.guid
LEFT JOIN locations l ON ip.location_guid = l.guid
WHERE ip.deleted_at IS NULL
  AND (CAST(sqlc.narg('search') AS text) IS NULL OR sqlc.narg('search') = '' OR ip.name ILIKE '%' || CAST(sqlc.narg('search') AS text) || '%' OR ip.notes ILIKE '%' || CAST(sqlc.narg('search') AS text) || '%')
  AND (CAST(sqlc.narg('device_guid') AS uuid) IS NULL OR ip.device_guid = sqlc.narg('device_guid'))
  AND (CAST(sqlc.narg('location_guid') AS uuid) IS NULL OR ip.location_guid = sqlc.narg('location_guid'))
ORDER BY ip.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountInstallationPoints :one
SELECT COUNT(*) FROM installation_points WHERE deleted_at IS NULL;

-- name: UpdateInstallationPoint :exec
UPDATE installation_points
SET name = $2, device_guid = $3, location_guid = $4, notes = $5, updated_at = $6
WHERE guid = $1 AND deleted_at IS NULL;

-- name: DeleteInstallationPoint :exec
UPDATE installation_points
SET deleted_at = $2, updated_at = $2
WHERE guid = $1 AND deleted_at IS NULL;

-- name: GetInstallationPointWithRelations :one
SELECT
    ip.guid, ip.name, ip.device_guid, ip.location_guid, ip.notes, ip.created_at, ip.updated_at, ip.deleted_at,
    d.serial_number, d.alias, l.name
FROM installation_points ip
LEFT JOIN devices d ON ip.device_guid = d.guid
LEFT JOIN locations l ON ip.location_guid = l.guid
WHERE ip.guid = $1 AND ip.deleted_at IS NULL;
