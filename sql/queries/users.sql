-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
	gen_random_uuid(),
	NOW(),
	NOW(),
	$1,
	$2
)
RETURNING *;

-- name: RemoveAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users SET email = $1, hashed_password = $2, updated_at = $3 WHERE id = $4 RETURNING *;

-- name: UpgradeUser :exec
UPDATE users SET is_chirpy_red = TRUE, updated_at = $1 WHERE id = $2;
