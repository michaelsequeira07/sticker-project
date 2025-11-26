package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

var (
	// dbPath can be overridden via DB_PATH environment variable
	dbPath = getEnv("DB_PATH", "stickers.db")
	db     *sql.DB
)

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Transaction represents an incoming transaction
type Transaction struct {
	TransactionID string `json:"transaction_id"`
	ShopperID     string `json:"shopper_id"`
	StoreID       string `json:"store_id"`
	Timestamp     string `json:"timestamp"`
	Items         []Item `json:"items"`
}

// Item represents an item in a transaction
type Item struct {
	SKU       string  `json:"sku"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Category  string  `json:"category"`
}

// TransactionRecord represents a stored transaction
type TransactionRecord struct {
	TransactionID  string  `json:"transaction_id"`
	ShopperID      string  `json:"shopper_id"`
	StoreID        string  `json:"store_id"`
	Timestamp      string  `json:"timestamp"`
	TotalAmount    float64 `json:"total_amount"`
	StickersEarned int     `json:"stickers_earned"`
	CreatedAt      string  `json:"created_at"`
	Items          []Item  `json:"items"`
}

// ShopperStatus represents a shopper's current status
type ShopperStatus struct {
	ShopperID      string              `json:"shopper_id"`
	CurrentBalance int                 `json:"current_balance"`
	TotalEarned    int                 `json:"total_earned"`
	TotalRedeemed  int                 `json:"total_redeemed"`
	Transactions   []TransactionRecord `json:"transactions"`
	Redemptions    []Redemption        `json:"redemptions"`
}

// Redemption represents a sticker redemption
type Redemption struct {
	ID           int    `json:"id"`
	ShopperID    string `json:"shopper_id"`
	RewardName   string `json:"reward_name"`
	StickersCost int    `json:"stickers_cost"`
	RedeemedAt   string `json:"redeemed_at"`
}

// RedemptionRequest represents a redemption request
type RedemptionRequest struct {
	ShopperID  string `json:"shopper_id"`
	RewardName string `json:"reward_name"`
}

// Stats represents system statistics
type Stats struct {
	TotalStickersAwarded  int          `json:"total_stickers_awarded"`
	TotalTransactions     int          `json:"total_transactions"`
	TotalRedemptions      int          `json:"total_redemptions"`
	TotalStickersRedeemed int          `json:"total_stickers_redeemed"`
	StickersPerStore      []StoreStats `json:"stickers_per_store"`
}

// StoreStats represents statistics per store
type StoreStats struct {
	StoreID          string `json:"store_id"`
	TotalStickers    int    `json:"total_stickers"`
	TransactionCount int    `json:"transaction_count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	queries := []string{
		`CREATE TABLE IF NOT EXISTS transactions (
			transaction_id TEXT PRIMARY KEY,
			shopper_id TEXT NOT NULL,
			store_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			total_amount REAL NOT NULL,
			stickers_earned INTEGER NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS transaction_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transaction_id TEXT NOT NULL,
			sku TEXT NOT NULL,
			name TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL,
			category TEXT NOT NULL,
			FOREIGN KEY (transaction_id) REFERENCES transactions(transaction_id)
		)`,
		`CREATE TABLE IF NOT EXISTS redemptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			shopper_id TEXT NOT NULL,
			reward_name TEXT NOT NULL,
			stickers_cost INTEGER NOT NULL,
			redeemed_at TEXT NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// CalculationBreakdown shows how stickers were calculated
type CalculationBreakdown struct {
	TotalAmount  float64 `json:"total_amount"`
	BaseStickers int     `json:"base_stickers"`
	PromoBonus   int     `json:"promo_bonus"`
	BeforeCap    int     `json:"before_cap"`
	AfterCap     int     `json:"after_cap"`
	Capped       bool    `json:"capped"`
}

// calculateStickers calculates stickers earned based on campaign rules
func calculateStickers(items []Item) int {
	breakdown := calculateStickersWithBreakdown(items)
	return breakdown.AfterCap
}

// calculateStickersWithBreakdown calculates stickers and returns breakdown
func calculateStickersWithBreakdown(items []Item) CalculationBreakdown {
	// Calculate total basket amount
	var totalAmount float64
	var promoBonus int

	for _, item := range items {
		totalAmount += float64(item.Quantity) * item.UnitPrice
		if item.Category == "promo" {
			promoBonus += item.Quantity
		}
	}

	// Base stickers: 1 per $10
	baseStickers := int(totalAmount / 10.0)

	// Total before cap
	totalStickers := baseStickers + promoBonus

	// Apply per-transaction cap of 5
	capped := totalStickers > 5
	afterCap := totalStickers
	if capped {
		afterCap = 5
	}

	return CalculationBreakdown{
		TotalAmount:  totalAmount,
		BaseStickers: baseStickers,
		PromoBonus:   promoBonus,
		BeforeCap:    totalStickers,
		AfterCap:     afterCap,
		Capped:       capped,
	}
}

// validateTransaction validates transaction input
func validateTransaction(tx *Transaction) error {
	if tx.TransactionID == "" {
		return fmt.Errorf("missing required field: transaction_id")
	}
	if tx.ShopperID == "" {
		return fmt.Errorf("missing required field: shopper_id")
	}
	if tx.StoreID == "" {
		return fmt.Errorf("missing required field: store_id")
	}
	if tx.Timestamp == "" {
		return fmt.Errorf("missing required field: timestamp")
	}
	if len(tx.Items) == 0 {
		return fmt.Errorf("items must be a non-empty list")
	}

	for _, item := range tx.Items {
		if item.SKU == "" {
			return fmt.Errorf("missing required item field: sku")
		}
		if item.Name == "" {
			return fmt.Errorf("missing required item field: name")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("item quantity must be positive: %s", item.SKU)
		}
		if item.UnitPrice < 0 {
			return fmt.Errorf("item unit_price cannot be negative: %s", item.SKU)
		}
		if item.Category == "" {
			return fmt.Errorf("missing required item field: category")
		}
	}

	return nil
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func createTransactionHandler(c *gin.Context) {
	var tx Transaction
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Request body must be valid JSON"})
		return
	}

	// Validate input
	if err := validateTransaction(&tx); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Check if transaction already exists (idempotency)
	var existingStickers int
	err := db.QueryRow(
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Calculate stickers and total amount with breakdown
	breakdown := calculateStickersWithBreakdown(tx.Items)
	stickersEarned := breakdown.AfterCap
	totalAmount := breakdown.TotalAmount

	// Log transaction processing
	log.Printf("[TRANSACTION] Processing transaction_id=%s shopper_id=%s store_id=%s total_amount=%.2f stickers_earned=%d (base=%d promo=%d before_cap=%d capped=%v)",
		tx.TransactionID, tx.ShopperID, tx.StoreID, totalAmount, stickersEarned,
		breakdown.BaseStickers, breakdown.PromoBonus, breakdown.BeforeCap, breakdown.Capped)

	// Begin transaction
	txDB, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
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
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
	}

	// Commit transaction
	if err = txDB.Commit(); err != nil {
		log.Printf("[TRANSACTION] Failed to commit transaction_id=%s: %v", tx.TransactionID, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	log.Printf("[TRANSACTION] Successfully processed transaction_id=%s shopper_id=%s stickers_earned=%d",
		tx.TransactionID, tx.ShopperID, stickersEarned)

	// Get current shopper balance
	var totalEarned int
	err = db.QueryRow(
		"SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions WHERE shopper_id = ?",
		tx.ShopperID,
	).Scan(&totalEarned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	var totalRedeemed int
	err = db.QueryRow(
		"SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions WHERE shopper_id = ?",
		tx.ShopperID,
	).Scan(&totalRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	currentBalance := totalEarned - totalRedeemed

	c.JSON(http.StatusCreated, gin.H{
		"transaction_id":  tx.TransactionID,
		"shopper_id":      tx.ShopperID,
		"stickers_earned": stickersEarned,
		"current_balance": currentBalance,
	})
}

func getShopperStatusHandler(c *gin.Context) {
	shopperID := c.Param("shopper_id")

	// Get all transactions
	rows, err := db.Query(
		`SELECT transaction_id, store_id, timestamp, total_amount, stickers_earned, created_at
		FROM transactions
		WHERE shopper_id = ?
		ORDER BY created_at DESC`,
		shopperID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	var transactions []TransactionRecord
	var totalEarned int

	for rows.Next() {
		var tr TransactionRecord
		if err := rows.Scan(&tr.TransactionID, &tr.StoreID, &tr.Timestamp, &tr.TotalAmount, &tr.StickersEarned, &tr.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		transactions = append(transactions, tr)
		totalEarned += tr.StickersEarned
	}

	// Get redemptions
	rows, err = db.Query(
		`SELECT reward_name, stickers_cost, redeemed_at
		FROM redemptions
		WHERE shopper_id = ?
		ORDER BY redeemed_at DESC`,
		shopperID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	var redemptions []Redemption
	var totalRedeemed int

	for rows.Next() {
		var r Redemption
		if err := rows.Scan(&r.RewardName, &r.StickersCost, &r.RedeemedAt); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		redemptions = append(redemptions, r)
		totalRedeemed += r.StickersCost
	}

	currentBalance := totalEarned - totalRedeemed

	status := ShopperStatus{
		ShopperID:      shopperID,
		CurrentBalance: currentBalance,
		TotalEarned:    totalEarned,
		TotalRedeemed:  totalRedeemed,
		Transactions:   transactions,
		Redemptions:    redemptions,
	}

	c.JSON(http.StatusOK, status)
}

func createRedemptionHandler(c *gin.Context) {
	var req RedemptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Request body must be valid JSON"})
		return
	}

	if req.ShopperID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "missing required field: shopper_id"})
		return
	}
	if req.RewardName == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "missing required field: reward_name"})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid reward. Available rewards: %v", available),
		})
		return
	}

	// Get current balance
	var totalEarned int
	err := db.QueryRow(
		"SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions WHERE shopper_id = ?",
		req.ShopperID,
	).Scan(&totalEarned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	var totalRedeemed int
	err = db.QueryRow(
		"SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions WHERE shopper_id = ?",
		req.ShopperID,
	).Scan(&totalRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	currentBalance := totalEarned - totalRedeemed

	if currentBalance < stickersCost {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Insufficient stickers. Current balance: %d, Required: %d", currentBalance, stickersCost),
		})
		return
	}

	// Create redemption
	redeemedAt := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(
		`INSERT INTO redemptions (shopper_id, reward_name, stickers_cost, redeemed_at)
		VALUES (?, ?, ?, ?)`,
		req.ShopperID, req.RewardName, stickersCost, redeemedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
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

func getStatsHandler(c *gin.Context) {
	var stats Stats

	// Total stickers awarded
	err := db.QueryRow("SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions").Scan(&stats.TotalStickersAwarded)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Total transactions
	err = db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&stats.TotalTransactions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Total redemptions
	err = db.QueryRow("SELECT COUNT(*) FROM redemptions").Scan(&stats.TotalRedemptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	err = db.QueryRow("SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions").Scan(&stats.TotalStickersRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Stickers per store
	rows, err := db.Query(
		`SELECT store_id, SUM(stickers_earned) as total_stickers, COUNT(*) as transaction_count
		FROM transactions
		GROUP BY store_id
		ORDER BY total_stickers DESC`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var ss StoreStats
		if err := rows.Scan(&ss.StoreID, &ss.TotalStickers, &ss.TransactionCount); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		stats.StickersPerStore = append(stats.StickersPerStore, ss)
	}

	c.JSON(http.StatusOK, stats)
}

func getTransactionDetailsHandler(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	// Get transaction
	var tr TransactionRecord
	err := db.QueryRow(
		`SELECT transaction_id, shopper_id, store_id, timestamp, total_amount, stickers_earned, created_at
		FROM transactions
		WHERE transaction_id = ?`,
		transactionID,
	).Scan(&tr.TransactionID, &tr.ShopperID, &tr.StoreID, &tr.Timestamp, &tr.TotalAmount, &tr.StickersEarned, &tr.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Transaction not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Get transaction items
	rows, err := db.Query(
		`SELECT sku, name, quantity, unit_price, category
		FROM transaction_items
		WHERE transaction_id = ?
		ORDER BY sku`,
		transactionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.SKU, &item.Name, &item.Quantity, &item.UnitPrice, &item.Category); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		tr.Items = append(tr.Items, item)
	}

	// Calculate breakdown for display
	breakdown := calculateStickersWithBreakdown(tr.Items)

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

// TransactionDebugInfo contains full debug information about a transaction
type TransactionDebugInfo struct {
	TransactionRecord
	Calculation    CalculationBreakdown `json:"calculation"`
	ShopperBalance int                  `json:"shopper_balance_after"`
	ProcessingTime string               `json:"processing_time"`
}

func getTransactionDebugHandler(c *gin.Context) {
	transactionID := c.Param("transaction_id")

	// Get transaction
	var tr TransactionRecord
	err := db.QueryRow(
		`SELECT transaction_id, shopper_id, store_id, timestamp, total_amount, stickers_earned, created_at
		FROM transactions
		WHERE transaction_id = ?`,
		transactionID,
	).Scan(&tr.TransactionID, &tr.ShopperID, &tr.StoreID, &tr.Timestamp, &tr.TotalAmount, &tr.StickersEarned, &tr.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Transaction not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Get transaction items
	rows, err := db.Query(
		`SELECT sku, name, quantity, unit_price, category
		FROM transaction_items
		WHERE transaction_id = ?
		ORDER BY sku`,
		transactionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.SKU, &item.Name, &item.Quantity, &item.UnitPrice, &item.Category); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
			return
		}
		tr.Items = append(tr.Items, item)
	}

	// Calculate breakdown
	breakdown := calculateStickersWithBreakdown(tr.Items)

	// Get shopper balance after this transaction
	var totalEarned int
	err = db.QueryRow(
		"SELECT COALESCE(SUM(stickers_earned), 0) FROM transactions WHERE shopper_id = ? AND created_at <= ?",
		tr.ShopperID, tr.CreatedAt,
	).Scan(&totalEarned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	var totalRedeemed int
	err = db.QueryRow(
		"SELECT COALESCE(SUM(stickers_cost), 0) FROM redemptions WHERE shopper_id = ? AND redeemed_at <= ?",
		tr.ShopperID, tr.CreatedAt,
	).Scan(&totalRedeemed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("Database error: %v", err)})
		return
	}

	shopperBalanceAfter := totalEarned - totalRedeemed

	debugInfo := TransactionDebugInfo{
		TransactionRecord: tr,
		Calculation:       breakdown,
		ShopperBalance:    shopperBalanceAfter,
		ProcessingTime:    tr.CreatedAt,
	}

	c.JSON(http.StatusOK, debugInfo)
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Set up router
	r := gin.Default()

	// Routes
	r.GET("/health", healthHandler)
	r.POST("/transactions", createTransactionHandler)
	r.GET("/shoppers/:shopper_id", getShopperStatusHandler)
	r.POST("/redemptions", createRedemptionHandler)
	r.GET("/stats", getStatsHandler)
	r.GET("/transactions/:transaction_id", getTransactionDetailsHandler)
	r.GET("/debug/transactions/:transaction_id", getTransactionDebugHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on http://0.0.0.0:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
