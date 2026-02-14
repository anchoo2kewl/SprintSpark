-- Add status_category to swim_lanes for reliable status mapping
-- Replaces fragile name-based matching with explicit category

ALTER TABLE swim_lanes ADD COLUMN status_category TEXT NOT NULL DEFAULT '';

-- Backfill existing swim lanes based on name patterns
UPDATE swim_lanes SET status_category = 'todo' WHERE LOWER(name) LIKE '%to do%' OR LOWER(name) LIKE '%todo%' OR LOWER(name) = 'backlog';
UPDATE swim_lanes SET status_category = 'in_progress' WHERE LOWER(name) LIKE '%progress%' OR LOWER(name) LIKE '%doing%' OR LOWER(name) LIKE '%review%';
UPDATE swim_lanes SET status_category = 'done' WHERE LOWER(name) LIKE '%done%' OR LOWER(name) LIKE '%complete%' OR LOWER(name) LIKE '%finished%';

-- Safety: default any unmatched to 'todo'
UPDATE swim_lanes SET status_category = 'todo' WHERE status_category = '';
