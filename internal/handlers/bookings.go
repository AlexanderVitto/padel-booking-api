package handlers

import (
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
)

type BookingsHandler struct{}

func NewBookingsHandler() *BookingsHandler {
	return &BookingsHandler{}
}

func (h *BookingsHandler) List(c *gin.Context) {
	response.OK(c, gin.H{"data": []any{}})
}
