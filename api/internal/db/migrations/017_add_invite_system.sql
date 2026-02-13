-- Add invite-only registration system
-- Each user gets 3 invites, admins get unlimited (enforced in code)

ALTER TABLE users ADD COLUMN invite_count INTEGER NOT NULL DEFAULT 3;

CREATE TABLE invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT UNIQUE NOT NULL,
    inviter_id INTEGER NOT NULL REFERENCES users(id),
    invitee_id INTEGER REFERENCES users(id),
    used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invites_code ON invites(code);
CREATE INDEX idx_invites_inviter_id ON invites(inviter_id);
