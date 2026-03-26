package queries

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CancelBooking membatalkan booking.
// requestUserID dipakai untuk ownership check — hanya pemilik yang boleh cancel.
func CancelBooking(ctx context.Context, pool *pgxpool.Pool, bookingID string, requestUserID string) (Booking, error) {
	// Step 1: coba update langsung
	// WHERE id=$1 AND user_id=$2 AND status='confirmed'
	// → hanya berhasil kalau booking ada, milik user ini, dan masih confirmed
	var b Booking
	err := pool.QueryRow(ctx, `
		update bookings
		set status = 'canceled'
		where id = $1 
			and user_id = $2
			and status = 'confirmed'
		returning id::text, court_id::text, user_id::text, start_time, end_time,
		status, created_at
		`, bookingID, requestUserID).Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt)

	if err == nil {
		// berhasil di-cancel
		return b, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		// error DB beneran
		return Booking{}, err
	}

	// Step 2: tidak ada row yang ke-update
	// cek apakah booking exists (mungkin sudah canceled)
	err = pool.QueryRow(ctx, `
		select id::text, court_id::text, user_id::text, start_time, end_time, status
		from bookings
		where id = $1
	`, bookingID).Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status)

	if errors.Is(err, pgx.ErrNoRows) {
		// booking tidak ada sama sekali
		return Booking{}, ErrBookingNotFound
	}

	// Step 2: update tidak berhasil → cari tahu kenapa
	// Kemungkinan:
	//   a) booking tidak ada → ErrBookingNotFound
	//   b) booking ada tapi bukan milik user ini → ErrForbidden
	//   c) booking ada, milik user ini, tapi sudah canceled → return booking (idempotent)
	err = pool.QueryRow(ctx, `
		select id::text, court_id::text, user_id::text, start_time, end_time, status,
		created_at
		from bookings
		where id = $1
	`, bookingID).Scan(
		&b.ID, &b.CourtID, &b.UserID,
		&b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		// (a) booking tidak ada sama sekali
		return Booking{}, ErrBookingNotFound
	}
	if err != nil {
		return Booking{}, err
	}

	// (b) booking ada tapi bukan milik user ini
	if b.UserID != requestUserID {
		return Booking{}, ErrForbidden
	}

	// (c) booking ada, milik user ini, tapi sudah canceled → idempotent
	return b, nil
}
