-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(user_id, token_hash, issued_at, expires_at, user_agent, ip)
VALUES(
     $1,
     $2,
     $3,
     $4,
     $5,
     $6
)
RETURNING user_id, token_hash, issued_at, expires_at, revoked_at, user_agent, ip;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, issued_at, expires_at, revoked_at, user_agent, ip
FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshTokenByID :exec
UPDATE refresh_tokens 
SET revoked_at = NOW()
WHERE id=$1 AND revoked_at IS NULL;

-- name: RevokeAllRefreshTokensForUser :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE expires_at < NOW();

-- name: GetActiveRefreshTokenByHash :one
SELECT id, user_id, token_hash, issued_at, expires_at, revoked_at, user_agent, ip
FROM refresh_tokens
WHERE token_hash = $1
AND revoked_at IS NULL
AND expires_at > NOW();

-- name: ListActiveTokensForUser :many
SELECT id, user_id, issued_at, expires_at, user_agent, ip
FROM refresh_tokens
WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > NOW()
ORDER BY issued_at DESC;