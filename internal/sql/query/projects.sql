-- name: ProjectShort :one
SELECT id, name, parent_id FROM projects WHERE id = ?;

-- name: ProjectDetail :one
select id, name, description, parent_id, planned_for, start_at, due_at, completed_at FROM projects WHERE id = ?;

-- name: ProjectName :one
SELECT name FROM projects WHERE id = ?;

-- name: RecProjectName :one
WITH RECURSIVE rec_project_name(id, name) AS (
    SELECT id, name FROM projects WHERE parent_id IS NULL
    UNION ALL
    SELECT projects.id, rec_project_name.name||' > '||projects.name FROM projects JOIN rec_project_name ON projects.parent_id = rec_project_name.id
) SELECT name FROM rec_project_name WHERE projects.id = ?;
