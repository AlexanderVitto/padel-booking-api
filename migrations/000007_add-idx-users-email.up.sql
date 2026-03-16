-- Mempercepat lookup saat login (WHERE email = $1)
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);