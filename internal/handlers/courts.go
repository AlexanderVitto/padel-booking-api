package handlers

import (
	"github.com/AlexanderVitto/padel-booking-api/internal/response"
	"github.com/gin-gonic/gin"
)

type CourtsHandler struct{}

func NewCourtsHandler() *CourtsHandler {
	return &CourtsHandler{}
}

func (h *CourtsHandler) List(c *gin.Context) {
	response.OK(c, gin.H{"data": []any{}})
}
