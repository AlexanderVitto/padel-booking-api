CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  text        NOT NULL UNIQUE,
    expires_at   timestamptz NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);

-- Mempercepat lookup saat refresh & logout (WHERE token_hash = $1)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash 
    ON refresh_tokens (token_hash);

-- Mempercepat lookup saat logout semua device (WHERE user_id = $1)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id
    ON refresh_tokens (user_id);