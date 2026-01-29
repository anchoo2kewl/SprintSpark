-- Add swim lanes table for customizable task statuses per project

-- Swim lanes define the columns in the kanban board for each project
CREATE TABLE swim_lanes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    color TEXT NOT NULL, -- hex color for the swim lane indicator
    position INTEGER NOT NULL, -- order of columns (0-based)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX idx_swim_lanes_project_id ON swim_lanes(project_id);
CREATE INDEX idx_swim_lanes_position ON swim_lanes(project_id, position);

-- Ensure unique positions per project
CREATE UNIQUE INDEX idx_swim_lanes_unique_position ON swim_lanes(project_id, position);

-- Trigger to update updated_at timestamp for swim_lanes
CREATE TRIGGER update_swim_lanes_timestamp
AFTER UPDATE ON swim_lanes
FOR EACH ROW
BEGIN
    UPDATE swim_lanes SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
END;

-- Update tasks table to reference swim_lane_id instead of hardcoded status
-- First, add the new column
ALTER TABLE tasks ADD COLUMN swim_lane_id INTEGER REFERENCES swim_lanes(id) ON DELETE SET NULL;

-- Create index for the new column
CREATE INDEX idx_tasks_swim_lane_id ON tasks(swim_lane_id);

-- Migrate existing projects to have default swim lanes
-- This will be done in a separate data migration after the schema is updated
