-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- USERS
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    display_name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- VENUES
CREATE TABLE IF NOT EXISTS venues(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    timezone text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- COURTS
CREATE TABLE IF NOT EXISTS courts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    venue_id uuid NOT NULL REFERENCES venues(id) ON DELETE CASCADE,
    name text NOT NULL,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(venue_id, name) -- Ensure court names are unique within a venue
);

CREATE INDEX IF NOT EXISTS idx_courts_venue_id ON courts(venue_id);

-- BOOKINGS
-- Slot is 60 minutes. We enforce:
-- - end_time = start_time + 60 minutes via CHECK constraint
-- - no double-booking for same court+start_time via UNIQUE index (confirmed only)
CREATE TABLE IF NOT EXISTS bookings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    court_id uuid NOT NULL REFERENCES courts(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'confirmed',
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT bookings_status_check CHECK (status IN ('confirmed','canceled')),
    CONSTRAINT bookings_duration_check CHECK (end_time = start_time + interval '60 minutes'),
    CONSTRAINT bookings_start_minute_check CHECK (date_part('minute', start_time) = 0 AND 
    date_part('second', start_time) = 0),
    CONSTRAINT bookings_end_minute_check CHECK (date_part('minute', end_time) = 0 AND
    date_part('second', end_time) = 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bookings_court_start_time
  ON bookings (court_id, start_time)
  WHERE status = 'confirmed';

CREATE INDEX IF NOT EXISTS idx_bookings_court_id_start_time ON bookings(court_id, start_time);
CREATE INDEX IF NOT EXISTS idx_bookings_user_id_created_at ON bookings(user_id, created_at);

-- COURT CHAT MESSAGES (room per court)
CREATE TABLE IF NOT EXISTS court_messages(
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    court_id uuid NOT NULL REFERENCES courts(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_court_messages_court_id_created_at ON court_messages(court_id, created_at);
CREATE INDEX IF NOT EXISTS idx_court_messages_user_id_created_at ON court_messages(user_id, created_at);