package handlers

import (
	"net/http"

	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	// Week 2: add DB ping here.
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
