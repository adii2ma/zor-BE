CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(18,2) NOT NULL CHECK (amount >= 0),
    type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
    category TEXT NOT NULL,
    transaction_date DATE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

--bun:split

CREATE INDEX IF NOT EXISTS transactions_user_id_idx ON transactions (user_id);

--bun:split

CREATE INDEX IF NOT EXISTS transactions_date_idx ON transactions (transaction_date);

--bun:split

CREATE INDEX IF NOT EXISTS transactions_category_idx ON transactions (category);

--bun:split

CREATE INDEX IF NOT EXISTS transactions_type_idx ON transactions (type);

--bun:split

CREATE INDEX IF NOT EXISTS transactions_user_date_idx ON transactions (user_id, transaction_date DESC);
