-- name: CreateSessionParticipant :one
INSERT INTO session_participant (user_id, session_id)
VALUES(
    $1,
    $2
)
RETURNING *;

-- name: DeleteSessionParticipant :exec
DELETE FROM session_participant
WHERE user_id = $1 AND session_id = $2;

-- name: GetSessionParticipant :one
SELECT *
FROM session_participant
WHERE user_id = $1 AND session_id = $2;

-- name: GetAllSessionParticipants :many
SELECT *
FROM session_participant
WHERE session_id = $1
ORDER BY joined_at DESC
LIMIT $2 OFFSET $3;

