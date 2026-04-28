-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), now(), now(), $1, $2)
RETURNING id, created_at, updated_at, email;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT id, created_at, updated_at, email, hashed_password
FROM users
WHERE email = $1;

-- name: UpdateUser :one
update users
set email           = $1,
    hashed_password = $2,
    updated_at      = now()
where id = $3
RETURNING id, created_at, updated_at, email;
