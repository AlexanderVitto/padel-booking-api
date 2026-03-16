package queries

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Booking struct {
	ID        string    `json:"id"`
	CourtID   string    `json:"court_id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type ListBookingsParams struct {
	CourtID string
	Date    time.Time
	Limit   int
	// Cursor adalah start_time booking terakhir dari halaman sebelumnya.
	// Nil → ambil dari awal hari.
	Cursor *time.Time
}

type ListBookingsResult struct {
	Bookings   []Booking  `json:"bookings"`
	NextCursor *time.Time `json:"next_cursor"`
	HasNext    bool       `json:"has_next"`
	Limit      int        `json:"limit"`
}

func ListBookingsByCourtAndDate(ctx context.Context, pool *pgxpool.Pool, p ListBookingsParams) (ListBookingsResult, error) {
	startOfDay := time.Date(p.Date.Year(), p.Date.Month(), p.Date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Ambil limit+1 untuk deteksi apakah ada halaman berikutnya.
	fetchLimit := p.Limit + 1

	var rows pgx.Rows
	var err error

	if p.Cursor == nil {
		// halaman pertama → mulai dari awal hari
		rows, err = pool.Query(ctx, `
			select id::text, court_id::text, user_id::text, start_time, end_time, status
			from bookings
			where court_id = $1
			  and start_time >= $2
			  and start_time < $3
			order by start_time asc
			limit $4
		`, p.CourtID, startOfDay, endOfDay, fetchLimit)
	} else {
		// halaman berikutnya → mulai dari setelah cursor
		rows, err = pool.Query(ctx, `
			select id::text, court_id::text, user_id::text, start_time, end_time, status
			from bookings
			where court_id = $1
			  and start_time > $2
			  and start_time < $3
			order by start_time asc
			limit $4
		`, p.CourtID, p.Cursor, endOfDay, fetchLimit)
	}
	if err != nil {
		return ListBookingsResult{}, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status); err != nil {
			return ListBookingsResult{}, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return ListBookingsResult{}, err
	}

	// hindari null di JSON kalau kosong
	if bookings == nil {
		bookings = []Booking{}
	}

	// cek apakah ada halaman berikutnya
	hasNext := len(bookings) > p.Limit
	if hasNext {
		// buang item ke limit+1, itu hanya penanda has_next
		bookings = bookings[:p.Limit]
	}

	// next_cursor = start_time booking terakhir di halaman ini
	var nextCursor *time.Time
	if hasNext && len(bookings) > 0 {
		t := bookings[len(bookings)-1].StartTime
		nextCursor = &t
	}

	return ListBookingsResult{
		Bookings:   bookings,
		NextCursor: nextCursor,
		HasNext:    hasNext,
		Limit:      p.Limit,
	}, nil
}

func GetBookingByID(ctx context.Context, pool *pgxpool.Pool, id string) (Booking, error) {
	var b Booking
	err := pool.QueryRow(ctx, `
		select id::text, court_id::text, user_id::text, start_time, end_time, status
		from bookings
		where id = $1
	`, id).Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status)
	if err != nil {
		if isNotFound(err) {
			return Booking{}, ErrBookingNotFound
		}
		return Booking{}, err
	}
	return b, nil
}
