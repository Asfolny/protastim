-- name: GetProject :one
SELECT * FROM projects WHERE id = ?;

-- name: ListProjects :many
SELECT * FROM projects;

-- name: CreateProject :one
INSERT INTO projects (name, description, status, start_at, due_at, budget, recurrence, parent_id, client_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET
  name = IIF(CAST(@set_name AS BOOLEAN), CAST(@name AS TEXT), name),
  description = IIF(CAST(@set_description AS BOOLEAN), CAST(sqlc.narg('description') AS TEXT), description),
  status = IIF(CAST(@set_status AS BOOLEAN), CAST(@status AS TEXT), status),
  start_at = IIF(CAST(@set_start_at AS BOOLEAN), CAST(sqlc.narg('start_at') AS DATETIME), start_at),
  due_at = IIF(CAST(@set_due_at AS BOOLEAN), CAST(sqlc.narg('due_at') AS DATETIME), due_at),
  budget = IIF(CAST(@set_budget AS BOOLEAN), CAST(sqlc.narg('budget') AS TEXT), budget),
  recurrence = IIF(CAST(@set_recurrence AS BOOLEAN), CAST(sqlc.narg('recurrence') AS TEXT), recurrence),
  parent_id = IIF(CAST(@set_parent AS BOOLEAN), CAST(sqlc.narg('parent_id') AS INTEGER), parent_id),
  client_id = IIF(CAST(@set_client AS BOOLEAN), CAST(sqlc.narg('client_id') AS INTEGER), client_id)
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;
