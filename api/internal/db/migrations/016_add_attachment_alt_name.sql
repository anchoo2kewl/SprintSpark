-- Add alt_name to task_attachments for searchable image library
ALTER TABLE task_attachments ADD COLUMN alt_name TEXT NOT NULL DEFAULT '';
