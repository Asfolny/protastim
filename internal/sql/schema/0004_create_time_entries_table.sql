-- +goose Up
CREATE TABLE time_entries (
  start_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  end_at DATETIME,
  task_id INTEGER NOT NULL REFERENCES tasks(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

-- +goose StatementBegin
CREATE TRIGGER time_entry_updated_at BEFORE UPDATE ON time_entries FOR EACH ROW
BEGIN
  UPDATE time_entries SET updated_at = CURRENT_TIMESTAMP WHERE rowid = NEW.rowid;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER time_entry_updated_at;
DROP TABLE time_entries;
