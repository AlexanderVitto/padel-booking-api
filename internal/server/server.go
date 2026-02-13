package server

import (
	"github.com/AlexanderVitto/padel-booking-api/internal/config"
	"github.com/AlexanderVitto/padel-booking-api/internal/handlers"
	"github.com/gin-gonic/gin"
)

func New(cfg config.Config) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// Middleware (week 1)
	r.Use(RequestID())
	r.Use(RequestLogger())
	r.Use(CORSDev())

	// Handlers
	healthHandler := handlers.NewHealthHandler()
	courtsHandler := handlers.NewCourtsHandler()
	bookingsHandler := handlers.NewBookingsHandler()

	// Routes
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	v1 := r.Group("/v1")
	{
		v1.GET("/courts", courtsHandler.List)
		v1.GET("/bookings", bookingsHandler.List)
	}

	return r

}
