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

	// Global middleware
	r.Use(RequestID())
	r.Use(RequestLogger())
	r.Use(CORSDev())

	// Handlers
	healthHandler := handlers.NewHealthHandler(pool)
	courtsHandler := handlers.NewCourtsHandler(pool)
	bookingsHandler := handlers.NewBookingsHandler(pool)
	pingHandler := handlers.NewPingHandler()
	authHandler := handlers.NewAuthHandler(pool, cfg)

	// Health (no auth)
	r.GET("/healthz", healthHandler.Healthz)
	r.GET("/readyz", healthHandler.Readyz)

	v1 := r.Group("/v1")
	{
		// Public routes (no authentication required)
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/refresh", authHandler.Refresh)
		v1.POST("/auth/logout", authHandler.Logout)

		// Protected routes (require JWT)
		protected := v1.Group("/")
		protected.Use(RequireJWT(cfg.JWTAccessSecret))
		{
			protected.GET("/ping", pingHandler.Ping)
			protected.GET("/courts", courtsHandler.List)
			protected.GET("/bookings", bookingsHandler.List)
			protected.GET("/bookings/:id", bookingsHandler.GetByID)
			protected.POST("/bookings", bookingsHandler.Create)
			protected.PATCH("/bookings/:id/cancel", bookingsHandler.Cancel)
		}
	}

	return r
}
