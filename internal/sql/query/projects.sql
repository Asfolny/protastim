-- name: GetProject :one
SELECT * FROM projects WHERE id = ?;

-- name: ListProjects :many
SELECT * FROM projects;

-- name: CreateProject :one
INSERT INTO projects (name, description, status)
VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET
  name = IIF(CAST(@set_name AS BOOLEAN), CAST(@name AS TEXT), name),
  description = IIF(CAST(@set_description AS BOOLEAN), CAST(sqlc.narg('description') AS TEXT), description),
  status = IIF(CAST(@set_status AS BOOLEAN), CAST(@status AS TEXT), status)
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;
