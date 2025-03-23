-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;

-- name: ListTasks :many
SELECT * FROM tasks;

-- name: CreateTask :one
INSERT INTO tasks (name, description, status, project_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET
  name = IIF(CAST(@set_name AS BOOLEAN), CAST(@name AS TEXT), name),
  description = IIF(CAST(@set_description AS BOOLEAN), CAST(sqlc.narg('description') AS TEXT), description),
  status = IIF(CAST(@set_status AS BOOLEAN), CAST(@status AS TEXT), status)
WHERE id = ?
RETURNING *;

-- name: MoveTask :one
UPDATE tasks
SET project_id = ?
WHERE id = ?
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ?;
