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

func (h *BookingsHandler) List(c *gin.Context) {
	courtID := c.Query("court_id")
	if courtID == "" {
		response.Error(c, http.StatusBadRequest, "validation_error", "court_id is required")
		return
	}
	if !isValidUUID(courtID) {
		response.Error(c, http.StatusBadRequest, "validation_error", "court_id must be a valid UUID")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		response.Error(c, http.StatusBadRequest, "validation_error", "date is required (YYYY-MM-DD)")
		return
	}
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "date must be in YYYY-MM-DD format")
		return
	}

	limit := parseIntQuery(c.Query("limit"), defaultLimit)
	if limit <= 0 || limit > maxLimit {
		response.Error(c, http.StatusBadRequest, "validation_error", "limit must be between 1 and 100")
		return
	}

	// parse cursor (opsional)
	var cursor *time.Time
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		t, err := time.Parse(time.RFC3339, cursorStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "validation_error", "cursor must be in RFC3339 format (e.g. 2026-02-28T10:00:00Z)")
			return
		}
		cursor = &t
	}

	result, err := queries.ListBookingsByCourtAndDate(c.Request.Context(), h.pool, queries.ListBookingsParams{
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

func (h *BookingsHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		response.Error(c, http.StatusBadRequest, "validation_error", "id must be a valid UUID")
		return
	}

	b, err := queries.GetBookingByID(c.Request.Context(), h.pool, id)
	if err != nil {
		if err == queries.ErrBookingNotFound {
			response.Error(c, http.StatusNotFound, "not_found", "booking not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to get booking")
		return
	}

	response.OK(c, gin.H{"booking": b})
}

type createBookingRequest struct {
	CourtID   string    `json:"court_id"   binding:"required"`
	UserID    string    `json:"user_id"    binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time"   binding:"required"`
}

func (h *BookingsHandler) Create(c *gin.Context) {
	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "invalid request body")
		return
	}

	if !isValidUUID(req.CourtID) {
		response.Error(c, http.StatusBadRequest, "validation_error", "court_id must be a valid UUID")
		return
	}
	if !isValidUUID(req.UserID) {
		response.Error(c, http.StatusBadRequest, "validation_error", "user_id must be a valid UUID")
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
		UserID:    req.UserID,
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

func (h *BookingsHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if !isValidUUID(id) {
		response.Error(c, http.StatusBadRequest, "validation_error", "id must be a valid UUID")
		return
	}

	b, err := queries.CancelBooking(c.Request.Context(), h.pool, id)
	if err != nil {
		if err == queries.ErrBookingNotFound {
			response.Error(c, http.StatusNotFound, "not_found", "booking not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to cancel booking")
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
