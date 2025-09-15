-- name: CreateSession :one
INSERT INTO session (session_code, session_name, creator_user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM session
WHERE id = $1;

-- name: GetActiveSessionByID :one 
SELECT * 
FROM session
WHERE id = $1;

-- name: GetUserSessions :many
SELECT *
FROM session
WHERE creator_user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: UpdateSessionName :exec
UPDATE session
SET session_name = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateSessionStatus :exec
UPDATE session
SET status = $2, updated_at = NOW()
WHERE id = $1;