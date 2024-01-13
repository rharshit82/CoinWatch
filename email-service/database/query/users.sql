-- name: GetUserEmailByAlertID :one
SELECT u.email
FROM "Users" u
INNER JOIN "Alerts" a ON u.id = a.user_id
WHERE a.id = $1;