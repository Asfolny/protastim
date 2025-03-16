-- +goose Up
CREATE TABLE projects (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'N' CHECK(status IN ('N', 'P', 'C', 'H', 'D')),
  start_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  due_at DATETIME,
  budget TEXT,
  recurrence TEXT,
  parent_id INTEGER REFERENCES projects(id),
  client_id INTEGER REFERENCES clients(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(name, client_id)
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
