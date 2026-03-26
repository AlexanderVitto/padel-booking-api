package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/AlexanderVitto/padel-booking-api/internal/db/queries"
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type BookingsHandler struct {
	pool *pgxpool.Pool
}

func NewBookingsHandler(pool *pgxpool.Pool) *BookingsHandler {
	return &BookingsHandler{pool: pool}
}

// mustGetUserID mengambil user_id dari JWT context yang di-set oleh RequireJWT middleware.
// Kalau tidak ada → return false dan langsung tulis 500 response.
// Ini seharusnya tidak pernah terjadi di protected routes.
func mustGetUserID(c *gin.Context) (string, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		response.Error(c, http.StatusInternalServerError, "internal_error", "missing user context")
		return "", false
	}
	userID, ok := v.(string)
	if !ok || userID == "" {
		response.Error(c, http.StatusInternalServerError, "internal_error", "invalid user context")
		return "", false
	}
	return userID, true
}

// List mengembalikan daftar booking milik user yang sedang login.
// Query params opsional: court_id, date (YYYY-MM-DD), limit, cursor (RFC3339)
func (h *BookingsHandler) List(c *gin.Context) {
	// ambil user_id dari JWT — user hanya bisa lihat booking miliknya
	userID, ok := mustGetUserID(c)
	if !ok {
		return
	}

	// court_id opsional
	var courtID *string
	if v := c.Query("court_id"); v != "" {
		if !isValidUUID(v) {
			response.Error(c, http.StatusBadRequest, "validation_error", "court_id must be a valid UUID")
			return
		}
		courtID = &v
	}

	// date opsional
	var date *time.Time
	if v := c.Query("date"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", "date must be in YYYY-MM-DD format")
			return
		}
		date = &t
	}

	limit := parseIntQuery(c.Query("limit"), defaultLimit)
	if limit <= 0 || limit > maxLimit {
		response.Error(c, http.StatusBadRequest, "validation_error", "limit must be between 1 and 100")
		return
	}

	// cursor opsional
	var cursor *time.Time
	if v := c.Query("cursor"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", "cursor must be in RFC3339 format (e.g. 2026-02-28T10:00:00Z)")
			return
		}
		cursor = &t
	}

	result, err := queries.ListBookingsByUser(c.Request.Context(), h.pool, queries.ListBookingsParams{
		UserID:  userID,
		CourtID: courtID,
		Date:    date,
		Limit:   limit,
		Cursor:  cursor,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list bookings")
		return
	}

	response.OK(c, result)
}

// GetByID mengembalikan booking by ID.
// Hanya pemilik booking yang bisa lihat.
func (h *BookingsHandler) GetByID(c *gin.Context) {
	userID, ok := mustGetUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if !isValidUUID(id) {
		response.Error(c, http.StatusBadRequest, "validation_error", "id must be a valid UUID")
		return
	}

	b, err := queries.GetBookingByID(c.Request.Context(), h.pool, id, userID)
	if err != nil {
		switch err {
		case queries.ErrBookingNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "booking not found")
		case queries.ErrForbidden:
			// jangan expose "booking exists but not yours" → pakai 404
			// ini mencegah user enumerate booking orang lain
			response.Error(c, http.StatusNotFound, "not_found", "booking not found")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get booking")
		}
		return
	}

	response.OK(c, gin.H{"booking": b})
}

// createBookingRequest — user_id TIDAK lagi ada di sini.
// user_id diambil dari JWT token, bukan dari body request.
type createBookingRequest struct {
	CourtID   string    `json:"court_id"   binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time"   binding:"required"`
}

// Create membuat booking baru.
// user_id diambil dari JWT — user hanya bisa booking atas nama dirinya sendiri.
func (h *BookingsHandler) Create(c *gin.Context) {
	userID, ok := mustGetUserID(c)
	if !ok {
		return
	}

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid request body")
		return
	}

	if !isValidUUID(req.CourtID) {
		response.Error(c, http.StatusBadRequest, "validation_error", "court_id must be a valid UUID")
		return
	}
	if !req.EndTime.After(req.StartTime) {
		response.Error(c, http.StatusBadRequest, "validation_error", "end_time must be after start_time")
		return
	}
	if req.StartTime.Before(time.Now().UTC()) {
		response.Error(c, http.StatusBadRequest, "validation_error", "start_time must be in the future")
		return
	}

	b, err := queries.CreateBooking(c.Request.Context(), h.pool, queries.CreateBookingParams{
		CourtID:   req.CourtID,
		UserID:    userID, // ← dari JWT, bukan dari body
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    "confirmed",
	})
	if err != nil {
		if err == queries.ErrBookingConflict {
			response.Error(c, http.StatusConflict, "booking_conflict", "time slot is already booked")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to create booking")
		return
	}

	response.Created(c, gin.H{"booking": b})
}

// Cancel membatalkan booking.
// Hanya pemilik booking yang boleh cancel.
func (h *BookingsHandler) Cancel(c *gin.Context) {
	userID, ok := mustGetUserID(c)
	if !ok {
		return
	}

	id := c.Param("id")
	if !isValidUUID(id) {
		response.Error(c, http.StatusBadRequest, "validation_error", "id must be a valid UUID")
		return
	}

	b, err := queries.CancelBooking(c.Request.Context(), h.pool, id, userID)
	if err != nil {
		switch err {
		case queries.ErrBookingNotFound:
			response.Error(c, http.StatusNotFound, "not_found", "booking not found")
		case queries.ErrForbidden:
			// sama seperti GetByID — pakai 404 bukan 403
			// mencegah user tau booking orang lain exists
			response.Error(c, http.StatusNotFound, "not_found", "booking not found")
		default:
			response.Error(c, http.StatusInternalServerError, "internal_error", "failed to cancel booking")
		}
		return
	}

	response.OK(c, gin.H{"booking": b})
}

func parseIntQuery(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}
