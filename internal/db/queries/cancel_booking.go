package queries

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CancelBooking(ctx context.Context, pool *pgxpool.Pool, bookingID string) (Booking, error) {
	// Step 1: coba update langsung (hanya kalau status masih confirmed)
	var b Booking
	err := pool.QueryRow(ctx, `
		update bookings
		set status = 'canceled'
		where id = $1
		  and status = 'confirmed'
		returning id::text, court_id::text, user_id::text, start_time, end_time, status
	`, bookingID).Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status)

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

	if err != nil {
		return Booking{}, err
	}

	// booking ada tapi sudah canceled → idempotent, return booking existing
	return b, nil
}
