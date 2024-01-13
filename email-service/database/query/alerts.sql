-- name: UpdateAlertStatus :exec
UPDATE "Alerts" SET
  status = $2
WHERE "id" = $1;