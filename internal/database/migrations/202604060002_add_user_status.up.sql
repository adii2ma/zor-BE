ALTER TABLE users
ADD COLUMN status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive'));

CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);
