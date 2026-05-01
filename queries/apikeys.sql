-- name: CreateAPIKey :exec
INSERT INTO apikeys (guid, key_hash, notes, expire_date, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetAPIKeyByGUID :one
SELECT guid, key_hash, notes, expire_date, is_active, created_at, updated_at, deleted_at
FROM apikeys
WHERE guid = $1 AND deleted_at IS NULL;

-- name: GetAPIKeyByKeyHash :one
SELECT guid, key_hash, notes, expire_date, is_active, created_at, updated_at, deleted_at
FROM apikeys
WHERE key_hash = $1 AND deleted_at IS NULL;

-- name: ListAPIKeys :many
SELECT guid, key_hash, notes, expire_date, is_active, created_at, updated_at, deleted_at
FROM apikeys
WHERE deleted_at IS NULL
  AND notes ILIKE $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAPIKeys :one
SELECT COUNT(*) FROM apikeys WHERE deleted_at IS NULL;

-- name: UpdateAPIKey :exec
UPDATE apikeys
SET notes = $2, expire_date = $3, is_active = $4, updated_at = $5
WHERE guid = $1 AND deleted_at IS NULL;

-- name: DeleteAPIKey :exec
UPDATE apikeys
SET deleted_at = $2, updated_at = $2
WHERE guid = $1 AND deleted_at IS NULL;
