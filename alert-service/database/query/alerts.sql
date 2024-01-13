-- name: CreateAlert :one
INSERT INTO "Alerts" (
  user_id, crypto, price, direction
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetAlertByID :one
SELECT * FROM "Alerts" 
WHERE "id" = $1;

-- name: GetAllAlerts :many
SELECT * FROM "Alerts" 
WHERE "user_id" = $1
LIMIT $2
OFFSET $3;

-- name: GetAlertsByStatus :many
SELECT * FROM "Alerts" 
WHERE "user_id" = $1 AND "status" = $2
LIMIT $3
OFFSET $4;

-- name: UpdateAlert :one
UPDATE "Alerts" SET
  crypto = $2,
  price = $3,
  direction = $4
WHERE "id" = $1
RETURNING *;

-- name: UpdateAlertStatus :exec
UPDATE "Alerts" SET
  status = $2
WHERE "id" = $1;