-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (guid, user_guid, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetRefreshTokenByTokenHash :one
SELECT guid, user_guid, token_hash, expires_at, created_at
FROM refresh_tokens
WHERE token_hash = $1;

-- name: DeleteRefreshTokensByUserGUID :exec
DELETE FROM refresh_tokens WHERE user_guid = $1;

-- name: DeleteRefreshTokenByTokenHash :exec
DELETE FROM refresh_tokens WHERE token_hash = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < $1;
