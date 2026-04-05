CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

--bun:split

INSERT INTO roles (id, name, description, created_at, updated_at)
VALUES
    ('viewer', 'Viewer', 'Can only view dashboard data', NOW(), NOW()),
    ('analyst', 'Analyst', 'Can view records and access insights', NOW(), NOW()),
    ('admin', 'Admin', 'Can create, update, and manage records and users', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

--bun:split

ALTER TABLE users
ADD COLUMN role_id TEXT NOT NULL DEFAULT 'viewer';

--bun:split

ALTER TABLE users
ADD CONSTRAINT users_role_id_fkey
FOREIGN KEY (role_id) REFERENCES roles(id);

--bun:split

CREATE INDEX IF NOT EXISTS users_role_id_idx ON users (role_id);
