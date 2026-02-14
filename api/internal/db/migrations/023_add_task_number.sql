-- Add per-project task numbering (like GitHub issues)
-- Each project's tasks are numbered sequentially starting from 1

ALTER TABLE tasks ADD COLUMN task_number INTEGER;

-- Backfill existing tasks with sequential numbers per project (ordered by creation)
-- SQLite doesn't support UPDATE with window functions directly, so we use a subquery
UPDATE tasks SET task_number = (
    SELECT cnt FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY id) AS cnt
        FROM tasks
    ) numbered
    WHERE numbered.id = tasks.id
);

-- Ensure task_number is unique per project
CREATE UNIQUE INDEX idx_tasks_project_task_number ON tasks(project_id, task_number);
