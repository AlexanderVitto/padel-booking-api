package queries

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// hashToken mengubah token string menjadi SHA-256 hex string.
// Ini yang disimpan di DB - bukan token aslinya.
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// SaveRefreshToken menyimpan hash dari refresh token ke DB.
func SaveRefreshToken(ctx context.Context, pool *pgxpool.Pool, userID string, token string, expiresAt time.Time) error {
	_, err := pool.Exec(ctx, `
		insert into refersh_tokens (user_id, token_hash, expires_at)
		values ($1, $2, $3)
	`, userID, hashToken(token), expiresAt)
	return err
}

// VerifyRefreshToken mengecek apakah token ada di DB dan belum expired.
// Return user_id kalau valid.
func VerifyRefreshToken(ctx context.Context, pool *pgxpool.Pool, token string) (string, error) {
	var userID string
	var expiresAt time.Time

	err := pool.QueryRow(ctx, `
		select user_id::text, expires_at
		from refresh_tokens
		where token_hash = $1
	`, hashToken(token)).Scan(&userID, &expiresAt)
	if err != nil {
		if isNotFound(err) {
			return "", ErrRefreshTokenNotFound
		}
		return "", err
	}

	// cek expired di aplikasi (double check selainn JWT expiry)
	if time.Now().After(expiresAt) {
		return "", ErrRefreshTokenExpired
	}
	return userID, nil
}

// DeleteRefreshToken menghapus token dari DB (logout).
func DeleteRefreshToken(ctx context.Context, pool *pgxpool.Pool, token string) error {
	_, err := pool.Exec(ctx, `
		delete from refresh_tokens
		where token_hash = $1
	`, hashToken(token))
	return err
}

// DeleteAllRefreshTokens menghapus semua token milik user (logout semua device).
func DeleteAllRefreshTokens(ctx context.Context, pool *pgxpool.Pool, userID string) error {
	_, err := pool.Exec(ctx, `
		delete from refresh_tokens
		where user_id = $1
	`, userID)
	return err
}
