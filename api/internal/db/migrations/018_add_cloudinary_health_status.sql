-- Add health monitoring columns to cloudinary_credentials
ALTER TABLE cloudinary_credentials ADD COLUMN status TEXT NOT NULL DEFAULT 'unknown';
ALTER TABLE cloudinary_credentials ADD COLUMN last_checked_at DATETIME;
ALTER TABLE cloudinary_credentials ADD COLUMN last_error TEXT NOT NULL DEFAULT '';
ALTER TABLE cloudinary_credentials ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0;
