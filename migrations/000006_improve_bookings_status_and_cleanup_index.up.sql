-- 1. Drop index duplikat dari migration 000005
--    (sudah ada idx_bookings_court_id_start_time di 000001)
DROP INDEX IF EXISTS idx_bookings_court_start;

-- 2. Drop check constraint lama
ALTER TABLE bookings
    DROP CONSTRAINT IF EXISTS bookings_status_check;

-- 3. Ubah tipe kolom status dari text ke varchar(20)
ALTER TABLE bookings
    ALTER COLUMN status TYPE varchar(20);

-- 4. Tambah check constraint baru dengan ejaan yang benar + pending
ALTER TABLE bookings
    ADD CONSTRAINT bookings_status_check
    CHECK (status IN ('confirmed', 'canceled', 'pending'));