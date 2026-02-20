package handlers

import (
	"net/http"

	"github.com/AlexanderVitto/padel-booking-api/internal/db/queries"
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourtsHandler struct {
	pool *pgxpool.Pool
}

func NewCourtsHandler(pool *pgxpool.Pool) *CourtsHandler {
	return &CourtsHandler{pool: pool}
}

func (h *CourtsHandler) List(c *gin.Context) {
	courts, err := queries.ListCourts(c.Request.Context(), h.pool)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", "failed to list courts")
		return
	}

	response.OK(c, gin.H{"courts": courts})
}
