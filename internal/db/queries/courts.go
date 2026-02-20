package queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Court struct {
	ID       string `json:"id"`
	VenueID  string `json:"venue_id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func ListCourts(ctx context.Context, pool *pgxpool.Pool) ([]Court, error) {
	rows, err := pool.Query(ctx, `
	select id::text, venue_id::text, name, is_active
	from courts
	where is_active = true
	order by created_at asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Court
	for rows.Next() {
		var c Court
		if err := rows.Scan(&c.ID, &c.VenueID, &c.Name, &c.IsActive); err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
