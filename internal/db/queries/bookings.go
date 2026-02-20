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
}

func ListBookingsByCourtAndDate(ctx context.Context, pool *pgxpool.Pool, courtID string, date time.Time) ([]Booking, error) {
	// Ambil semua booking pada tannggal itu (00:00 sampai sebelum besok 00:00)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	rows, err := pool.Query(ctx, `
	select id::text, court_id::text, user_id::text, start_time, end_time, status
	from bookings
	where court_id = $1
	  and start_time >= $2
	  and start_time < $3
	order by start_time asc
	`, courtID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.CourtID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
