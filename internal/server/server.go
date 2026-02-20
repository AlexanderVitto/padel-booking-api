package server

import (
	"github.com/AlexanderVitto/padel-booking-api/internal/config"
	"github.com/AlexanderVitto/padel-booking-api/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, pool *pgxpool.Pool) *gin.Engine {
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
	healthHandler := handlers.NewHealthHandler(pool)
	courtsHandler := handlers.NewCourtsHandler(pool)
	bookingsHandler := handlers.NewBookingsHandler(pool)
	pingHandler := handlers.NewPingHandler()

	// Routes
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	v1 := r.Group("/v1")
	v1.Use(RequireAPIKeyForV1())
	{
		v1.GET("/ping", pingHandler.Ping)
		v1.GET("/courts", courtsHandler.List)
		v1.GET("/bookings", bookingsHandler.List)
	}

	return r

}
