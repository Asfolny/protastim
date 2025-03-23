-- +goose Up
CREATE TABLE tasks (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'N' CHECK(status IN ('N', 'P', 'C', 'H', 'D')),
  project_id INTEGER NOT NULL REFERENCES projects(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE TRIGGER task_updated_at BEFORE UPDATE ON tasks FOR EACH ROW
BEGIN
  UPDATE tasks SET updated_at = CURRENT_TIMESTAMP WHERE rowid = NEW.rowid;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER task_updated_at;
DROP TABLE tasks;
