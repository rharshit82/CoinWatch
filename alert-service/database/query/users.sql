-- name: CreateUser :one
INSERT INTO "Users" (
  email, hashed_password
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetUserById :one
select * from "Users"
where id = $1
limit 1;

-- name: GetUserByEmail :one
select * from "Users"
where email = $1;