-- name: CreateUser :exec
INSERT INTO users (guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetUserByGUID :one
SELECT guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at, deleted_at
FROM users
WHERE guid = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT guid, email, password_hash, first_name, last_name, role_guid, is_active, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
  AND (email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- name: UpdateUser :exec
UPDATE users
SET email = $2, password_hash = $3, first_name = $4, last_name = $5, role_guid = $6, is_active = $7, updated_at = $8
WHERE guid = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = $2, updated_at = $2
WHERE guid = $1 AND deleted_at IS NULL;

-- name: ExistsUserByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL);

-- name: UpdateUserLastLogin :exec
UPDATE users SET updated_at = $2 WHERE guid = $1;
