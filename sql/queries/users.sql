-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: DeleteUsers :exec
DELETE FROM users;
