-- name: CreateThreshold :one
INSERT INTO thresholds (
    device_guid, parameter_name, min_value, max_value, severity, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetThresholdByGUID :one
SELECT * FROM thresholds
WHERE guid = $1 AND deleted_at IS NULL LIMIT 1;

-- name: ListThresholdsByDevice :many
SELECT * FROM thresholds
WHERE device_guid = $1 AND deleted_at IS NULL
ORDER BY parameter_name ASC;

-- name: UpdateThreshold :one
UPDATE thresholds
SET 
    parameter_name = $2,
    min_value = $3,
    max_value = $4,
    severity = $5,
    is_active = $6,
    updated_at = NOW()
WHERE guid = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteThreshold :exec
UPDATE thresholds
SET deleted_at = NOW()
WHERE guid = $1;
