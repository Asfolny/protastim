-- name: ScheduledTasks :many
WITH RECURSIVE rec_project_name(id, name, level) AS (
  SELECT id, name, 1 AS level FROM projects WHERE parent_id IS NULL
  UNION ALL
  SELECT
    projects.id,
    rec_project_name.name||' > '||projects.name,
    rec_project_name.level + 1
  FROM projects
  JOIN rec_project_name ON projects.parent_id = rec_project_name.id
)
SELECT tasks.id, tasks.name, project_id, planned_for, start_at, due_at, rec_project_name.name AS project_name, rec_project_name.level AS project_level
FROM tasks
JOIN rec_project_name ON tasks.project_id = rec_project_name.id
WHERE planned_for <= DATE() AND completed_at IS NULL ORDER BY planned_for, start_at NULLS LAST, due_at NULLS LAST;

-- name: GetTasksByProject :many
SELECT * FROM tasks WHERE project_id = ?;

-- name: StartTask :exec
UPDATE tasks SET start_at = IIF(start_at IS NULL, DATE(), start_at) WHERE id = ?;

-- name: CompleteTask :exec
UPDATE tasks SET completed_at = IIF(completed_at IS NULL, DATE(), completed_at) WHERE id = ?;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = ?;
