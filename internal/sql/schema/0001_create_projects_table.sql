-- +goose Up
CREATE TABLE projects (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  parent_id INTEGER REFERENCES projects(id),
  planned_for DATE,
  start_at DATE,
  due_at DATE,
  completed_at DATE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX proj_name ON projects (name);

-- +goose StatementBegin
CREATE TRIGGER project_updated_at AFTER UPDATE ON projects FOR EACH ROW
BEGIN
  UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER project_updated_at;
