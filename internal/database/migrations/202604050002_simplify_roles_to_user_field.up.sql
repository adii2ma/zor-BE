ALTER TABLE users
ADD COLUMN role TEXT NOT NULL DEFAULT 'viewer';

--bun:split

UPDATE users
SET role = role_id
WHERE role_id IN ('viewer', 'analyst', 'admin');

--bun:split

ALTER TABLE users
ADD CONSTRAINT users_role_check
CHECK (role IN ('viewer', 'analyst', 'admin'));

--bun:split

CREATE INDEX IF NOT EXISTS users_role_idx ON users (role);

--bun:split

DROP INDEX IF EXISTS users_role_id_idx;

--bun:split

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_id_fkey;

--bun:split

ALTER TABLE users
DROP COLUMN IF EXISTS role_id;

--bun:split

DROP TABLE IF EXISTS roles;
