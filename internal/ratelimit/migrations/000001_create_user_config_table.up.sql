CREATE TABLE IF NOT EXISTS user_configs (
ip text PRIMARY KEY,
capacity int NOT NULL,
rate_per_sec float NOT NULL,
updated_at TIMESTAMPTZ DEFAULT NOW()
);

