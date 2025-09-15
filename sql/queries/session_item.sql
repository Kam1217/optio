-- name: CreateSessionItem :one 
INSERT INTO session_item (
    session_id, 
    item_title, 
    item_description,
    image_url,
    source_type,
    source_id,
    metadata,
    added_by_user_id
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: DeleteSessionItem :exec
DELETE FROM session_item
WHERE id = $1;

-- name: GetSessionItemByID :one
SELECT *
FROM session_item
WHERE id = $1;

-- name: ListSessionItems :many
SELECT *
FROM session_item
WHERE session_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: UpdateItemTitle :exec
UPDATE session_item
SET item_title = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateItemDescription :exec
UPDATE session_item
SET item_description = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateItemImage :exec
UPDATE session_item
SET image_url = $2, updated_at = NOW()
WHERE id = $1;