-- name: StartTimeTracking :exec
INSERT INTO time_entries (task_id, start_at) VALUES (?, ?);

-- name: StopTimeTracking :exec
UPDATE time_entries SET end_at = ? WHERE end_at IS NULL;

-- name: GetRunningTimer :one
SELECT start_at, task_id FROM time_entries WHERE end_at IS NULL;

-- name: EnterTimeEntry :one
INSERT INTO time_entries (task_id, start_at, end_at) VALUES (?, ?, ?)
RETURNING *;

-- name: UpdateTimeEntry :one
UPDATE time_entries SET task_id = ?, start_at = ?, end_at = ? WHERE id = ?
RETURNING *;

-- name: ListTimeEntries :many
SELECT start_at, end_at, task_id FROM time_entries WHERE ? BETWEEN start_at AND due_at;

-- name: TimeByTask :many
SELECT start_at, end_at FROM time_entries WHERE task_id = ?;

-- name: WorkingOnWithName :one
SELECT time_entries.start_at, time_entries.task_id, tasks.name
FROM time_entries
JOIN tasks ON tasks.id = time_entries.task_id
WHERE end_at IS NULL;
