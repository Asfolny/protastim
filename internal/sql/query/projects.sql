-- name: ProjectShort :one
SELECT id, name, parent_id FROM projects WHERE id = ?;

-- name: GetProject :one
select * FROM projects WHERE id = ?;

-- name: ProjectName :one
SELECT name FROM projects WHERE id = ?;

-- name: RecProjectName :one
WITH RECURSIVE rec_project_name(id, name) AS (
    SELECT id, name FROM projects WHERE parent_id IS NULL
    UNION ALL
    SELECT projects.id, rec_project_name.name||' > '||projects.name FROM projects JOIN rec_project_name ON projects.parent_id = rec_project_name.id
) SELECT name FROM rec_project_name WHERE projects.id = ?;

-- name: ProjectsWithRecName :many
WITH RECURSIVE rec_project_name(id, name) AS (
    SELECT id, name FROM projects WHERE parent_id IS NULL
    UNION ALL
    SELECT projects.id, rec_project_name.name||' > '||projects.name FROM projects JOIN rec_project_name ON projects.parent_id = rec_project_name.id
) SELECT
  projects.id,
  projects.name,
  projects.description,
  projects.parent_id,
  projects.planned_for,
  projects.start_at,
  projects.due_at,
  projects.completed_at,
  rec_project_name.name AS parent_tree
FROM projects
LEFT JOIN rec_project_name ON projects.parent_id = rec_project_name.id
ORDER BY COALESCE(parent_id, projects.id), projects.id;

-- name: StartProject :exec
UPDATE projects SET start_at = IIF(start_at IS NULL, DATE(), start_at) WHERE id = ?;

-- name: CompleteProject :exec
UPDATE projects SET completed_at = IIF(completed_at IS NULL, DATE(), start_at) WHERE id = ?;

-- name: TopProjects :many
SELECT id, name FROM projects WHERE parent_id IS NULL;

-- name: CreateProject :one
INSERT INTO projects (name, description, parent_id, planned_for, start_at, due_at, completed_at) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateProjectName :exec
UPDATE projects SET name = ? WHERE id = ?;

-- name: UpdateProjectDescription :exec
UPDATE projects SET description = ? WHERE id = ?;

-- name: UpdateProjectParent :exec
UPDATE projects SET parent_id = ? WHERE id = ?;

-- name: UpdateProjectPlannedFor :exec
UPDATE projects SET planned_for = ? WHERE id = ?;

-- name: UpdateProjectStartAt :exec
UPDATE projects SET start_at = ? WHERE id = ?;

-- name: UpdateProjectDueAt :exec
UPDATE projects SET due_at = ? WHERE id = ?;

-- name: UpdateProjectCompletedAt :exec
UPDATE projects SET completed_at = ? WHERE id = ?;
