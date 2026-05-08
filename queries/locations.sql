-- name: CreateLocation :exec
INSERT INTO locations (guid, name, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLocationByGUID :one
SELECT l.guid, l.name, l.notes, l.created_at, l.updated_at, l.deleted_at,
       (SELECT COUNT(*) FROM installation_points ip WHERE ip.location_guid = l.guid AND ip.deleted_at IS NULL) as device_count
FROM locations l
WHERE l.guid = $1 AND l.deleted_at IS NULL;

-- name: ListLocations :many
SELECT l.guid, l.name, l.notes, l.created_at, l.updated_at, l.deleted_at,
       (SELECT COUNT(*) FROM installation_points ip WHERE ip.location_guid = l.guid AND ip.deleted_at IS NULL) as device_count
FROM locations l
WHERE l.deleted_at IS NULL
  AND (l.name ILIKE $1 OR l.notes ILIKE $1)
ORDER BY l.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountLocations :one
SELECT COUNT(*) FROM locations WHERE deleted_at IS NULL;

-- name: UpdateLocation :exec
UPDATE locations
SET name = $2, notes = $3, updated_at = $4
WHERE guid = $1 AND deleted_at IS NULL;

-- name: DeleteLocation :exec
UPDATE locations
SET deleted_at = $2, updated_at = $2
WHERE guid = $1 AND deleted_at IS NULL;

-- name: ExistsLocationByName :one
SELECT EXISTS(SELECT 1 FROM locations WHERE name = $1 AND deleted_at IS NULL);
