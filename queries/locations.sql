-- name: CreateLocation :exec
INSERT INTO locations (guid, name, notes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLocationByGUID :one
SELECT guid, name, notes, created_at, updated_at, deleted_at
FROM locations
WHERE guid = $1 AND deleted_at IS NULL;

-- name: ListLocations :many
SELECT guid, name, notes, created_at, updated_at, deleted_at
FROM locations
WHERE deleted_at IS NULL
  AND (name ILIKE $1 OR notes ILIKE $1)
ORDER BY created_at DESC
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
