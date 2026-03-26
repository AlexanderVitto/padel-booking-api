package queries

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Booking struct {
	ID        string    `json:"id"`
	CourtID   string    `json:"court_id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ListBookingsParams struct {
	// UserID wajib — user hanya bisa lihat booking miliknya sendiri
	UserID string
	// CourtID opsional — kalau diisi, filter by court
	CourtID *string
	// Date opsional — kalau diisi, filter by tanggal
	Date  *time.Time
	Limit int
	// Cursor adalah start_time booking terakhir dari halaman sebelumnya.
	// Nil → ambil dari awal.
	Cursor *time.Time
}

type ListBookingsResult struct {
	Bookings   []Booking  `json:"bookings"`
	NextCursor *time.Time `json:"next_cursor"`
	HasNext    bool       `json:"has_next"`
	Limit      int        `json:"limit"`
}

func ListBookingsByUser(ctx context.Context, pool *pgxpool.Pool, p ListBookingsParams) (ListBookingsResult, error) {
	// fetchLimit+1 untuk deteksi apakah ada halaman berikutnya
	fetchLimit := p.Limit + 1

	// base query — selalu filter by user_id
	// filter court_id & date kalau diisi
	// cursor-based pagination by start_time
	query := `
		select id::text, court_id::text, user_id::text, start_time, end_time, status, created_at
		from bookings
		where user_id = $1
		  and ($2::uuid is null or court_id = $2::uuid)
		  and ($3::date is null or start_time::date = $3::date)
		  and ($4::timestamptz is null or start_time > $4)
		order by start_time asc
		limit $5
	`

	rows, err := pool.Query(ctx, query,
		p.UserID,
		p.CourtID, // nil → $2 is null → tidak filter court
		p.Date,    // nil → $3 is null → tidak filter tanggal
		p.Cursor,  // nil → $4 is null → mulai dari awal
		fetchLimit,
	)
	if err != nil {
		return ListBookingsResult{}, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(
			&b.ID, &b.CourtID, &b.UserID,
			&b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt,
		); err != nil {
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

// GetBookingByID mengambil booking by ID.
// requestUserID dipakai untuk ownership check:
//   - kalau booking bukan milik requestUserID → ErrForbidden
//   - kalau booking tidak ada → ErrBookingNotFound
func GetBookingByID(ctx context.Context, pool *pgxpool.Pool, id string, requestUserID string) (Booking, error) {
	var b Booking
	err := pool.QueryRow(ctx, `
		select id::text, court_id::text, user_id::text, start_time, end_time, status, created_at
		from bookings
		where id = $1
	`, id).Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt)
	if err != nil {
		if isNotFound(err) {
			return Booking{}, ErrBookingNotFound
		}
		return Booking{}, err
	}

	// ownership check — booking ada tapi bukan milik user ini
	if b.UserID != requestUserID {
		return Booking{}, ErrForbidden
	}

	return b, nil
}
