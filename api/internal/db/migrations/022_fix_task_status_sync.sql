-- Fix tasks where status doesn't match their swim lane's status_category
-- This is a one-time data fix for existing out-of-sync tasks

UPDATE tasks SET status = (
    SELECT sl.status_category FROM swim_lanes sl WHERE sl.id = tasks.swim_lane_id
)
WHERE swim_lane_id IS NOT NULL
AND status != (SELECT sl.status_category FROM swim_lanes sl WHERE sl.id = tasks.swim_lane_id);
