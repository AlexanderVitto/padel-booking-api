package queries

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrBookingConflict      = errors.New("booking conflict")
	ErrBookingNotFound      = errors.New("booking not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailTaken           = errors.New("email already taken")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

// isNotFound adalah helper internal untuk cek pgx.ErrNoRows.
func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
