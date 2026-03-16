-- Rollback: kembalikann ke kondisi semula

-- 1. Drop check constraint baru
ALTER TABLE bookings
    DROP CONSTRAINT IF EXISTS bookings_status_check;

-- 2. Kembalikan tipe kolom status ke text
ALTER TABLE bookings
    ALTER COLUMN status TYPE text;

-- 3. Kembalikan check constraint lama
ALTER TABLE bookings
    ADD CONSTRAINT bookings_status_check 
    CHECK (status IN ('confirmed', 'canceled'));

-- 4. Kembalikan index yang di-drop
CREATE INDEX IF NOT EXISTS idx_bookings_court_start
    ON bookings (court_id, start_time);