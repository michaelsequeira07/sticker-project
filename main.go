package main

import (
	"log"

	"github.com/gin-gonic/gin"
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

	// Set up router
	r := gin.Default()

	// Routes
	r.GET("/health", handlers.HealthHandler)
	r.POST("/transactions", handlers.CreateTransactionHandler)
	r.GET("/shoppers/:shopper_id", handlers.GetShopperStatusHandler)
	r.POST("/redemptions", handlers.CreateRedemptionHandler)
	r.GET("/stats", handlers.GetStatsHandler)
	r.GET("/transactions/:transaction_id", handlers.GetTransactionDetailsHandler)
	r.GET("/debug/transactions/:transaction_id", handlers.GetTransactionDebugHandler)

	port := config.GetEnv("PORT", "8080")

	log.Printf("Starting server on http://0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
