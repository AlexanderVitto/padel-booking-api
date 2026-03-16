package queries

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateBookingParams struct {
	CourtID   string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

func CreateBooking(ctx context.Context, pool *pgxpool.Pool, p CreateBookingParams) (Booking, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Booking{}, err
	}
	defer tx.Rollback(ctx)

	var b Booking
	err = tx.QueryRow(ctx, `
	    insert into bookings (court_id, user_id, start_time, end_time, status)
		values ($1, $2, $3, $4, $5)
		returning id::text, court_id::text, user_id::text, start_time, end_time, status
	`, p.CourtID, p.UserID, p.StartTime, p.EndTime, p.Status).Scan(
		&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status,
	)
	if err != nil {
		// Deteksi konflik dari constraint/exclusion constraint Postgres.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// 23P01 adalah kode error untuk exclusion_violation, yang terjadi saat ada konflik dengan booking yang sudah ada.
			if pgErr.Code == "23P01" { // exclusion_violation
				return Booking{}, ErrBookingConflict
			}
			// 23505 adalah kode error untuk unique_violation, yang bisa terjadi jika ada constraint unik pada tabel (misalnya jika kita menambahkan constraint unik pada court_id + start_time).
			if pgErr.Code == "23505" { // unique_violation
				return Booking{}, ErrBookingConflict
			}
		}
		return Booking{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Booking{}, err
	}

	return b, nil
}
