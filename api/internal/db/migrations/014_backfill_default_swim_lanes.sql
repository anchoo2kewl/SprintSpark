-- Backfill default swim lanes for all existing projects

-- For each project, create the default swim lanes (To Do, In Progress, Done)
INSERT INTO swim_lanes (project_id, name, color, position)
SELECT
    p.id,
    'To Do',
    '#6B7280', -- gray
    0
FROM projects p;

INSERT INTO swim_lanes (project_id, name, color, position)
SELECT
    p.id,
    'In Progress',
    '#3B82F6', -- blue
    1
FROM projects p;

INSERT INTO swim_lanes (project_id, name, color, position)
SELECT
    p.id,
    'Done',
    '#10B981', -- green
    2
FROM projects p;

-- Update existing tasks to map old status values to swim_lane_id
-- Map 'todo' status to the 'To Do' swim lane
UPDATE tasks
SET swim_lane_id = (
    SELECT sl.id
    FROM swim_lanes sl
    WHERE sl.project_id = tasks.project_id
    AND sl.name = 'To Do'
    LIMIT 1
)
WHERE status = 'todo';

-- Map 'in_progress' status to the 'In Progress' swim lane
UPDATE tasks
SET swim_lane_id = (
    SELECT sl.id
    FROM swim_lanes sl
    WHERE sl.project_id = tasks.project_id
    AND sl.name = 'In Progress'
    LIMIT 1
)
WHERE status = 'in_progress';

-- Map 'done' status to the 'Done' swim lane
UPDATE tasks
SET swim_lane_id = (
    SELECT sl.id
    FROM swim_lanes sl
    WHERE sl.project_id = tasks.project_id
    AND sl.name = 'Done'
    LIMIT 1
)
WHERE status = 'done';

-- Note: We keep the 'status' column for now to maintain backward compatibility
-- It will be deprecated in a future migration once all clients are updated
