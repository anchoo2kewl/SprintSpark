ALTER TABLE tasks ADD COLUMN updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_updated_by ON tasks(updated_by);
