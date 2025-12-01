package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michaelsequeira07/sticker-project/auth"
	"github.com/michaelsequeira07/sticker-project/cache"
	"github.com/michaelsequeira07/sticker-project/calculator"
	"github.com/michaelsequeira07/sticker-project/database"
	"github.com/michaelsequeira07/sticker-project/models"
	"github.com/michaelsequeira07/sticker-project/validation"
)

// HealthHandler handles health check requests
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// CreateTransactionHandler handles transaction creation
func CreateTransactionHandler(c *gin.Context) {
	var tx models.Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Request body must be valid JSON"})
		return
	}

	// Validate input
	if err := validation.ValidateTransaction(&tx); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Check if transaction already exists (idempotency)
	var existingStickers int
	err := database.DB.QueryRow(
		"SELECT stickers_earned FROM transactions WHERE transaction_id = ?",
		tx.TransactionID,
	).Scan(&existingStickers)

	if err == nil {
		// Transaction already exists
		log.Printf("[TRANSACTION] Duplicate transaction_id=%s detected, returning existing result (stickers_earned=%d)",
			tx.TransactionID, existingStickers)
		c.JSON(http.StatusOK, gin.H{
			"message":         "Transaction already processed",
			"transaction_id":  tx.TransactionID,
			"stickers_earned": existingStickers,
		})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Calculate stickers and total amount with breakdown
	breakdown := calculator.CalculateStickersWithBreakdown(tx.Items)
	stickersEarned := breakdown.AfterCap
	totalAmount := breakdown.TotalAmount

	// Log transaction processing
	log.Printf("[TRANSACTION] Processing transaction_id=%s shopper_id=%s store_id=%s total_amount=%.2f stickers_earned=%d (base=%d promo=%d before_cap=%d capped=%v)",
		tx.TransactionID, tx.ShopperID, tx.StoreID, totalAmount, stickersEarned,
		breakdown.BaseStickers, breakdown.PromoBonus, breakdown.BeforeCap, breakdown.Capped)

	// Begin transaction
	txDB, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer txDB.Rollback()

	// Insert transaction
	createdAt := time.Now().UTC().Format(time.RFC3339)
	_, err = txDB.Exec(
		`INSERT INTO transactions 
		(transaction_id, shopper_id, store_id, timestamp, total_amount, stickers_earned, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tx.TransactionID, tx.ShopperID, tx.StoreID, tx.Timestamp, totalAmount, stickersEarned, createdAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Insert transaction items
	for _, item := range tx.Items {
		_, err = txDB.Exec(
			`INSERT INTO transaction_items 
			(transaction_id, sku, name, quantity, unit_price, category)
			VALUES (?, ?, ?, ?, ?, ?)`,
			tx.TransactionID, item.SKU, item.Name, item.Quantity, item.UnitPrice, item.Category,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
	}

	// Commit transaction
	if err = txDB.Commit(); err != nil {
		log.Printf("[TRANSACTION] Failed to commit transaction_id=%s: %v", tx.TransactionID, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Invalidate cache after new transaction
	if cache.Client != nil {
		cache.InvalidateShopperBalance(tx.ShopperID)
	}

	log.Printf("[TRANSACTION] Successfully processed transaction_id=%s shopper_id=%s stickers_earned=%d",
		tx.TransactionID, tx.ShopperID, stickersEarned)

	// Get current shopper balance (with Redis caching)
	currentBalance, err := getShopperBalanceWithCache(tx.ShopperID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"transaction_id":  tx.TransactionID,
		"shopper_id":      tx.ShopperID,
		"stickers_earned": stickersEarned,
		"current_balance": currentBalance,
	})
}

// GetShopperStatusHandler handles shopper status requests
func GetShopperStatusHandler(c *gin.Context) {
	shopperID := c.Param("shopper_id")

	// Get all transactions
	rows, err := database.DB.Query(
		`SELECT transaction_id, store_id, timestamp, total_amount, stickers_earned, created_at
		FROM transactions
		WHERE shopper_id = ?
		ORDER BY created_at DESC`,
		shopperID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	var transactions []models.TransactionRecord
	var totalEarned int

	for rows.Next() {
		var tr models.TransactionRecord
		if err := rows.Scan(&tr.TransactionID, &tr.StoreID, &tr.Timestamp, &tr.TotalAmount, &tr.StickersEarned, &tr.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		transactions = append(transactions, tr)
		totalEarned += tr.StickersEarned
	}

	// Get redemptions
	rows, err = database.DB.Query(
		`SELECT reward_name, stickers_cost, redeemed_at
		FROM redemptions
		WHERE shopper_id = ?
		ORDER BY redeemed_at DESC`,
		shopperID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	var redemptions []models.Redemption
	var totalRedeemed int

	for rows.Next() {
		var r models.Redemption
		if err := rows.Scan(&r.RewardName, &r.StickersCost, &r.RedeemedAt); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		redemptions = append(redemptions, r)
		totalRedeemed += r.StickersCost
	}

	// Use cached balance if available, otherwise calculate
	currentBalance, err := getShopperBalanceWithCache(shopperID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	status := models.ShopperStatus{
		ShopperID:      shopperID,
		CurrentBalance: currentBalance,
		TotalEarned:    totalEarned,
		TotalRedeemed:  totalRedeemed,
		Transactions:   transactions,
		Redemptions:    redemptions,
	}

	c.JSON(http.StatusOK, status)
}

// CreateRedemptionHandler handles redemption requests
func CreateRedemptionHandler(c *gin.Context) {
	var req models.RedemptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Request body must be valid JSON"})
		return
	}

	if err := validation.ValidateRedemptionRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Define available rewards
	rewards := map[string]int{
		"Mug":      10,
		"Tote bag": 20,
	}

	stickersCost, exists := rewards[req.RewardName]
	if !exists {
		available := make([]string, 0, len(rewards))
		for k := range rewards {
			available = append(available, k)
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("Invalid reward. Available rewards: %v", available),
		})
		return
	}

	// Get current balance
	var totalEarned int
	err := database.DB.QueryRow(
		"SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions WHERE shopper_id = ?",
		req.ShopperID,
	).Scan(&totalEarned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	var totalRedeemed int
	err = database.DB.QueryRow(
		"SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions WHERE shopper_id = ?",
		req.ShopperID,
	).Scan(&totalRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	currentBalance := totalEarned - totalRedeemed

	if currentBalance < stickersCost {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: fmt.Sprintf("Insufficient stickers. Current balance: %d, Required: %d", currentBalance, stickersCost),
		})
		return
	}

	// Create redemption
	redeemedAt := time.Now().UTC().Format(time.RFC3339)
	result, err := database.DB.Exec(
		`INSERT INTO redemptions (shopper_id, reward_name, stickers_cost, redeemed_at)
		VALUES (?, ?, ?, ?)`,
		req.ShopperID, req.RewardName, stickersCost, redeemedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Invalidate cache after redemption
	if cache.Client != nil {
		cache.InvalidateShopperBalance(req.ShopperID)
	}

	id, _ := result.LastInsertId()
	newBalance := currentBalance - stickersCost

	c.JSON(http.StatusCreated, gin.H{
		"id":               id,
		"shopper_id":       req.ShopperID,
		"reward_name":      req.RewardName,
		"stickers_cost":    stickersCost,
		"previous_balance": currentBalance,
		"new_balance":      newBalance,
		"redeemed_at":      redeemedAt,
	})
}

// GetStatsHandler handles statistics requests
func GetStatsHandler(c *gin.Context) {
	var stats models.Stats

	// Total stickers awarded
	err := database.DB.QueryRow("SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions").Scan(&stats.TotalStickersAwarded)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Total transactions
	err = database.DB.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&stats.TotalTransactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Total redemptions
	err = database.DB.QueryRow("SELECT COUNT(*) FROM redemptions").Scan(&stats.TotalRedemptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	err = database.DB.QueryRow("SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions").Scan(&stats.TotalStickersRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Stickers per store
	rows, err := database.DB.Query(
		`SELECT store_id, SUM(stickers_earned) as total_stickers, COUNT(*) as transaction_count
		FROM transactions
		GROUP BY store_id
		ORDER BY total_stickers DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ss models.StoreStats
		if err := rows.Scan(&ss.StoreID, &ss.TotalStickers, &ss.TransactionCount); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		stats.StickersPerStore = append(stats.StickersPerStore, ss)
	}

	c.JSON(http.StatusOK, stats)
}

// getShopperBalanceWithCache gets shopper balance with Redis caching
func getShopperBalanceWithCache(shopperID string) (int, error) {
	// Try to get from cache first
	if cache.Client != nil {
		balance, found, err := cache.GetShopperBalance(shopperID)
		if err == nil && found {
			return balance, nil
		}
	}

	// Cache miss - calculate from database
	var totalEarned int
	err := database.DB.QueryRow(
		"SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions WHERE shopper_id = ?",
		shopperID,
	).Scan(&totalEarned)
	if err != nil {
		return 0, err
	}

	var totalRedeemed int
	err = database.DB.QueryRow(
		"SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions WHERE shopper_id = ?",
		shopperID,
	).Scan(&totalRedeemed)
	if err != nil {
		return 0, err
	}

	currentBalance := totalEarned - totalRedeemed

	// Cache the result for 5 minutes
	if cache.Client != nil {
		cache.SetShopperBalance(shopperID, currentBalance, 5*time.Minute)
	}

	return currentBalance, nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
}

// LoginHandler handles user login and returns JWT token
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Request body must be valid JSON"})
		return
	}

	// Simple authentication (in production, verify against database)
	// For demo purposes, accept any username/password
	// In production, you would:
	// 1. Look up user in database
	// 2. Verify password hash
	// 3. Generate token

	token, err := auth.GenerateToken(req.Username, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"username": req.Username,
		"user_id":  req.Username,
	})
}

// RegisterHandler handles user registration
func RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Request body must be valid JSON"})
		return
	}

	// In production, you would:
	// 1. Hash the password
	// 2. Store user in database
	// 3. Generate token

	token, err := auth.GenerateToken(req.UserID, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":    token,
		"username": req.Username,
		"user_id":  req.UserID,
		"message":  "User registered successfully",
	})
}

// GetTransactionDetailsHandler handles transaction details requests
func GetTransactionDetailsHandler(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	// Get transaction
	var tr models.TransactionRecord
	err := database.DB.QueryRow(
		`SELECT transaction_id, shopper_id, store_id, timestamp, total_amount, stickers_earned, created_at
		FROM transactions
		WHERE transaction_id = ?`,
		transactionID,
	).Scan(&tr.TransactionID, &tr.ShopperID, &tr.StoreID, &tr.Timestamp, &tr.TotalAmount, &tr.StickersEarned, &tr.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Transaction not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Get transaction items
	rows, err := database.DB.Query(
		`SELECT sku, name, quantity, unit_price, category
		FROM transaction_items
		WHERE transaction_id = ?
		ORDER BY sku`,
		transactionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.SKU, &item.Name, &item.Quantity, &item.UnitPrice, &item.Category); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		tr.Items = append(tr.Items, item)
	}

	// Calculate breakdown for display
	breakdown := calculator.CalculateStickersWithBreakdown(tr.Items)

	// Enhanced response with calculation breakdown
	response := gin.H{
		"transaction_id":  tr.TransactionID,
		"shopper_id":      tr.ShopperID,
		"store_id":        tr.StoreID,
		"timestamp":       tr.Timestamp,
		"total_amount":    tr.TotalAmount,
		"stickers_earned": tr.StickersEarned,
		"created_at":      tr.CreatedAt,
		"items":           tr.Items,
		"calculation":     breakdown,
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionDebugHandler handles debug transaction requests
func GetTransactionDebugHandler(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	// Get transaction
	var tr models.TransactionRecord
	err := database.DB.QueryRow(
		`SELECT transaction_id, shopper_id, store_id, timestamp, total_amount, stickers_earned, created_at
		FROM transactions
		WHERE transaction_id = ?`,
		transactionID,
	).Scan(&tr.TransactionID, &tr.ShopperID, &tr.StoreID, &tr.Timestamp, &tr.TotalAmount, &tr.StickersEarned, &tr.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Transaction not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Get transaction items
	rows, err := database.DB.Query(
		`SELECT sku, name, quantity, unit_price, category
		FROM transaction_items
		WHERE transaction_id = ?
		ORDER BY sku`,
		transactionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.SKU, &item.Name, &item.Quantity, &item.UnitPrice, &item.Category); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		tr.Items = append(tr.Items, item)
	}

	// Calculate breakdown
	breakdown := calculator.CalculateStickersWithBreakdown(tr.Items)

	// Get shopper balance after this transaction
	var totalEarned int
	err = database.DB.QueryRow(
		"SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions WHERE shopper_id = ? AND created_at <= ?",
		tr.ShopperID, tr.CreatedAt,
	).Scan(&totalEarned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	var totalRedeemed int
	err = database.DB.QueryRow(
		"SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions WHERE shopper_id = ? AND redeemed_at <= ?",
		tr.ShopperID, tr.CreatedAt,
	).Scan(&totalRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	shopperBalanceAfter := totalEarned - totalRedeemed

	debugInfo := models.TransactionDebugInfo{
		TransactionRecord: tr,
		Calculation:       breakdown,
		ShopperBalance:    shopperBalanceAfter,
		ProcessingTime:    tr.CreatedAt,
	}

	c.JSON(http.StatusOK, debugInfo)
}
