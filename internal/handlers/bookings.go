package handlers

import (
	"net/http"
	"time"

	"github.com/AlexanderVitto/padel-booking-api/internal/db/queries"
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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

	dateStr := c.Query("date")
	if dateStr == "" {
		response.Error(c, http.StatusBadRequest, "validation_error", "date is required")
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "validation_error", "date must be in YYYY-MM-DD format")
		return
	}

	bookings, err := queries.ListBookingsByCourtAndDate(c.Request.Context(), h.pool, courtID, date)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list bookings")
		return
	}

	response.OK(c, gin.H{"bookings": bookings})
}
