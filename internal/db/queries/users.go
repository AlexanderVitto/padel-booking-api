package queries

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"` // jangan pernah expose ke JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserParams struct {
	Email       string
	Password    string // plain text → di-hash di sini
	DisplayName string
}

func CreateUser(ctx context.Context, pool *pgxpool.Pool, p CreateUserParams) (User, error) {
	// hash password sebelum simpan ke DB
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	var u User
	err = pool.QueryRow(ctx, `
		insert into users (email, password_hash, display_name)
		values ($1, $2, $3)
		returning id::text, email, display_name, created_at, updated_at
	`, p.Email, string(hash), p.DisplayName).Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}

	return u, nil
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (User, error) {
	var u User
	err := pool.QueryRow(ctx, `
		select id::text, email, password_hash, display_name, created_at, updated_at
		from users
		where email = $1
	`, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}

// GetUserByID dipakai saat refresh token - ambil data user terbaru dari DB.
func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id string) (User, error) {
	var u User
	err := pool.QueryRow(ctx, `
		select id::text, email, display_name, created_at, updated_at
		from users
		where id = $1
	`, id).Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}

func CheckPassword(user User, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainPassword))
}

type UpdateUserParams struct {
	ID string
	// DisplayName opsional - nil berarti tidak diupdate
	DisplayName *string
	// NewPassword opsional — nil berarti tidak diupdate
	// kalau diisi → di-hash sebelum disimpan
	NewPassword *string
}

// UpdateUser melakukan partial update pada user.
// Hanya field yang tidak nil yang akan diupdate.
func UpdateUser(ctx context.Context, pool *pgxpool.Pool, p UpdateUserParams) (User, error) {
	// hash password baru kalau diisi
	var newPasswordHash *string
	if p.NewPassword != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*p.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		h := string(hash)
		newPasswordHash = &h
	}

	// COALESCE trick:
	// COALESCE($2, display_name) → kalau $2 nil, pakai nilai lama (tidak berubah)
	//                            → kalau $2 ada, pakai nilai baru
	var u User
	err := pool.QueryRow(ctx, `
		update users
		set display_name = coalesce($2, display_name),
		    password_hash = coalesce($3, password_hash),
		    updated_at = now()
		where id = $1
		returning id::text, email, display_name, created_at, updated_at
	`, p.ID, p.DisplayName, newPasswordHash).Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}

// GetUserWithHashByID dipakai untuk verifikasi password — include password_hash.
func GetUserWithHashByID(ctx context.Context, pool *pgxpool.Pool, id string) (User, error) {
	var u User
	err := pool.QueryRow(ctx, `
		select id::text, email, password_hash, display_name, created_at, updated_at
		from users
		where id = $1
	`, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}

	return u, nil
}
