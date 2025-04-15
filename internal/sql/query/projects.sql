-- name: GetProject :one
SELECT * FROM projects WHERE id = ?;

-- name: ListProjects :many
SELECT * FROM projects;

-- name: ListProjectsStarted :many
SELECT *
FROM projects
WHERE start_at <= DATE() AND (completed_at IS NULL OR completed_at >= date('now', '-7 days'))
ORDER BY due_at;

-- name: ListProjectsOverdue :many
SELECT * FROM projects WHERE due_at <= DATE() AND completed_at IS NULL;

-- name: CreateProject :one
INSERT INTO projects (name, description, start_at, due_at, completed_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET
  name = ?,
  description = ?,
  start_at = ?,
  due_at = ?,
  completed_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;
