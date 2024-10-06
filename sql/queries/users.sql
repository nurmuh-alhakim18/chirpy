-- name: CreateUser :one
INSERT INTO users (id, email, created_at, updated_at, hashed_password)
VALUES (
  gen_random_uuid(), $1, NOW(), NOW(), $2
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, created_at, updated_at, expires_at, revoked_at)
VALUES (
  $1, $2, NOW(), NOW(), $3, NULL
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = NOW(), updated_at = NOW()
WHERE token = $1
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT u.*
FROM users AS u
JOIN refresh_tokens AS r ON u.id = r.user_id
WHERE r.token = $1
  AND r.revoked_at IS NULL
  AND r.expires_at > NOW();

-- name: UpgradeChirpyRed :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;
