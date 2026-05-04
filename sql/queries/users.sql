-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), now(), now(), $1, $2)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUser :one
SELECT id, created_at, updated_at, email, hashed_password, is_chirpy_red
FROM users
WHERE email = $1;

-- name: UpdateUser :one
update users
set email           = $1,
    hashed_password = $2,
    updated_at      = now()
where id = $3
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: UpgradeUser :one
update users
set is_chirpy_red = true,
    updated_at    = now()
where id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;
