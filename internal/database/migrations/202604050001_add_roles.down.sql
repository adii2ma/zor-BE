DROP INDEX IF EXISTS users_role_id_idx;

--bun:split

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_role_id_fkey;

--bun:split

ALTER TABLE users
DROP COLUMN IF EXISTS role_id;

--bun:split

DROP TABLE IF EXISTS roles;
