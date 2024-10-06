-- name: CreateChirp :one
INSERT INTO chirps (id, user_id, body, created_at, updated_at)
VALUES (
  gen_random_uuid(), $1, $2, NOW(), NOW()
)
RETURNING *;

-- name: GetAllChirps :many
SELECT id, user_id, body, created_at, updated_at
FROM chirps
ORDER BY created_at;

-- name: GetChirpById :one
SELECT id, user_id, body, created_at, updated_at
FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;