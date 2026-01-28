-- Add project_members table for granular project access control
-- This allows team owners to grant specific projects to team members
CREATE TABLE project_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member', -- 'owner', 'editor', 'member', 'viewer'
    granted_by INTEGER NOT NULL, -- user_id who granted access
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(project_id, user_id)
);

CREATE INDEX idx_project_members_project_id ON project_members(project_id);
CREATE INDEX idx_project_members_user_id ON project_members(user_id);
CREATE INDEX idx_project_members_role ON project_members(role);

-- Migrate existing data: Add project owners as members with 'owner' role
INSERT INTO project_members (project_id, user_id, role, granted_by, granted_at)
SELECT
    p.id as project_id,
    p.owner_id as user_id,
    'owner' as role,
    p.owner_id as granted_by,
    p.created_at as granted_at
FROM projects p;
