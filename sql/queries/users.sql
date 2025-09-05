-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, username, email, created_at, updated_at, deleted_at;

-- name: GetUserByUsername :one
SELECT id, username, email, created_at, updated_at, password_changed_at, deleted_at
FROM users
WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, username, email, created_at, updated_at, password_changed_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, username, email, created_at, updated_at, password_changed_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: UserExistsByUsernameOrEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE (username = $1 OR email = $2) AND deleted_at IS NULL);

-- name: GetUserForLogin :one
SELECT id, username, email, password_hash, password_changed_at, deleted_at
FROM users
WHERE (username = $1 OR email = $1) AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at 
FROM users 
WHERE deleted_at IS NULL
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, password_changed_at= NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUsername :exec
UPDATE users
SET username = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateEmail :exec
UPDATE users
SET email = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;