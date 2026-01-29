-- Migration: Backfill project access for existing team members
-- This migration adds all team members to their team's existing projects

-- First, check if granted_by and granted_at columns exist, if not add them
-- Note: SQLite doesn't support conditional ALTER, so we use a workaround

-- Create temporary table with new schema
CREATE TABLE IF NOT EXISTS project_members_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    granted_by INTEGER NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(project_id, user_id)
);

-- Copy existing data if project_members exists
INSERT OR IGNORE INTO project_members_new (project_id, user_id, role, granted_by, granted_at)
SELECT
    project_id,
    user_id,
    COALESCE(role, 'member'),
    COALESCE((SELECT owner_id FROM projects WHERE id = project_id LIMIT 1), user_id),
    CURRENT_TIMESTAMP
FROM project_members;

-- Drop old table and rename new one
DROP TABLE IF EXISTS project_members;
ALTER TABLE project_members_new RENAME TO project_members;

-- Recreate indices
CREATE INDEX IF NOT EXISTS idx_project_members_project_id ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_project_members_role ON project_members(role);

-- Add team members to all projects in their team (excluding members who already have access)
INSERT OR IGNORE INTO project_members (project_id, user_id, role, granted_by, granted_at)
SELECT DISTINCT p.id, tm.user_id, 'member', p.owner_id, CURRENT_TIMESTAMP
FROM projects p
INNER JOIN team_members tm ON p.team_id = tm.team_id
WHERE tm.status = 'active'
  AND NOT EXISTS (
    SELECT 1 FROM project_members pm
    WHERE pm.project_id = p.id AND pm.user_id = tm.user_id
  );
