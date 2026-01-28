-- Add name field to users table
ALTER TABLE users ADD COLUMN name TEXT;

-- Update existing user
UPDATE users SET name = 'Anshuman Biswas' WHERE email = 'anchoo2kewl@gmail.com';
