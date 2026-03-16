CREATE UNIQUE INDEX IF NOT EXISTS bookings_unique_confirmed_court_start
ON bookings (court_id, start_time)
WHERE status = 'confirmed';