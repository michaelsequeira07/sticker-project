package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/michaelsequeira07/sticker-project/auth"
	"github.com/michaelsequeira07/sticker-project/cache"
	"github.com/michaelsequeira07/sticker-project/config"
	"github.com/michaelsequeira07/sticker-project/database"
	"github.com/michaelsequeira07/sticker-project/handlers"
)

func main() {
	// Initialize database
	dbPath := config.GetEnv("DB_PATH", "stickers.db")
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.DB.Close()

	// Initialize Redis cache (optional - will continue if Redis unavailable)
	redisURL := config.GetEnv("REDIS_URL", "localhost:6379")
	if err := cache.InitCache(redisURL); err != nil {
		log.Printf("Warning: Redis cache unavailable: %v. Continuing without cache.", err)
	} else {
		log.Println("Redis cache initialized successfully")
	}

	// Set JWT secret from environment
	jwtSecret := config.GetEnv("JWT_SECRET", "your-secret-key-change-in-production")
	auth.SetJWTSecret(jwtSecret)

	// Set up router
	r := gin.Default()

	// Public routes (no authentication required)
	r.GET("/health", handlers.HealthHandler)
	r.POST("/login", handlers.LoginHandler)
	r.POST("/register", handlers.RegisterHandler)

	// Protected routes (require JWT authentication)
	protected := r.Group("/")
	protected.Use(auth.JWTAuthMiddleware())
	{
		protected.POST("/transactions", handlers.CreateTransactionHandler)
		protected.GET("/shoppers/:shopper_id", handlers.GetShopperStatusHandler)
		protected.POST("/redemptions", handlers.CreateRedemptionHandler)
		protected.GET("/stats", handlers.GetStatsHandler)
		protected.GET("/transactions/:transaction_id", handlers.GetTransactionDetailsHandler)
		protected.GET("/debug/transactions/:transaction_id", handlers.GetTransactionDebugHandler)
	}

	port := config.GetEnv("PORT", "8080")

	log.Printf("Starting server on http://0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
