-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, email, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT id, username, password_hash, email, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, username, password_hash, email, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UserExistsByUsername :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at 
FROM users 
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;