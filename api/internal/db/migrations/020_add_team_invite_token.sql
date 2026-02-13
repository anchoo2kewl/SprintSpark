-- Add acceptance token and expiry to team_invitations for one-click email acceptance
ALTER TABLE team_invitations ADD COLUMN acceptance_token TEXT;
ALTER TABLE team_invitations ADD COLUMN token_expires_at TIMESTAMP;
ALTER TABLE team_invitations ADD COLUMN invite_code TEXT;

CREATE UNIQUE INDEX idx_team_invitations_acceptance_token ON team_invitations(acceptance_token);
