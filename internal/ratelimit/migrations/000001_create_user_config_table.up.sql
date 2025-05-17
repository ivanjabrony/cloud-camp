CREATE TABLE IF NOT EXISTS user_configs (
ip varchar() PRIMARY KEY,
capacity int NOT NULL,
rate_per_sec float64 NOT NULL,
updated_at TIMESTAMPTZ DEFAULT NOW()
);

