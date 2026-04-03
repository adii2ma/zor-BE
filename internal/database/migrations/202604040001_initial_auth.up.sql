CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    google_subject TEXT NOT NULL,
    email TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    name TEXT NOT NULL,
    given_name TEXT,
    family_name TEXT,
    picture TEXT,
    locale TEXT,
    hosted_domain TEXT,
    google_issuer TEXT NOT NULL,
    google_authorized_party TEXT,
    google_audience TEXT NOT NULL,
    google_issued_at TIMESTAMPTZ NOT NULL,
    google_expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    last_login_at TIMESTAMPTZ NOT NULL
);

--bun:split

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    provider TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NOT NULL,
    user_agent TEXT,
    ip_address TEXT
);

--bun:split

CREATE UNIQUE INDEX IF NOT EXISTS users_google_subject_uidx ON users (google_subject);

--bun:split

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uidx ON users (email);

--bun:split

CREATE UNIQUE INDEX IF NOT EXISTS sessions_token_uidx ON sessions (token);

--bun:split

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
