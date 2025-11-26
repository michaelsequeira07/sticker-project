package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const testDBPath = "test_stickers.db"

func setupTestDB(t *testing.T) {
	// Remove test database if it exists
	os.Remove(testDBPath)
	
	// Close existing connection if any
	if db != nil {
		db.Close()
	}
	
	// Use test database
	dbPath = testDBPath
	
	// Initialize database
	if err := initDB(); err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

func teardownTestDB(t *testing.T) {
	if db != nil {
		db.Close()
		db = nil
	}
	os.Remove(testDBPath)
	// Reset to default
	dbPath = "stickers.db"
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", healthHandler)
	r.POST("/transactions", createTransactionHandler)
	r.GET("/shoppers/:shopper_id", getShopperStatusHandler)
	r.POST("/redemptions", createRedemptionHandler)
	r.GET("/stats", getStatsHandler)
	r.GET("/transactions/:transaction_id", getTransactionDetailsHandler)
	return r
}

func TestCalculateStickers(t *testing.T) {
	tests := []struct {
		name     string
		items    []Item
		expected int
	}{
		{
			name: "Base earn rate - $10",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 10, Category: "grocery"},
			},
			expected: 1,
		},
		{
			name: "Base earn rate - $19",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 19, Category: "grocery"},
			},
			expected: 1,
		},
		{
			name: "Base earn rate - $21",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 21, Category: "grocery"},
			},
			expected: 2,
		},
		{
			name: "Promo bonus",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 10, Category: "grocery"},
				{SKU: "SKU-2", Name: "Promo Item", Quantity: 1, UnitPrice: 5, Category: "promo"},
			},
			expected: 2, // $15 = 1 base, +1 promo = 2
		},
		{
			name: "Multiple promo items",
			items: []Item{
				{SKU: "SKU-1", Name: "Promo Item", Quantity: 2, UnitPrice: 5, Category: "promo"},
			},
			expected: 3, // $10 = 1 base, +2 promo = 3
		},
		{
			name: "Per-transaction cap",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 100, Category: "grocery"},
			},
			expected: 5, // $100 = 10 base, but capped at 5
		},
		{
			name: "Cap with promo",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 50, Category: "grocery"},
				{SKU: "SKU-2", Name: "Promo Item", Quantity: 3, UnitPrice: 5, Category: "promo"},
			},
			expected: 5, // $65 = 6 base, +3 promo = 9, but capped at 5
		},
		{
			name: "Zero amount",
			items: []Item{
				{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 0, Category: "grocery"},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateStickers(tt.items)
			assert.Equal(t, tt.expected, result, "calculateStickers() = %v, want %v", result, tt.expected)
		})
	}
}

func TestValidateTransaction(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
	}{
		{
			name: "Valid transaction",
			tx: Transaction{
				TransactionID: "tx-1",
				ShopperID:     "shopper-1",
				StoreID:       "store-1",
				Timestamp:     "2025-01-10T10:15:00Z",
				Items: []Item{
					{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 10, Category: "grocery"},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing transaction_id",
			tx: Transaction{
				ShopperID: "shopper-1",
				StoreID:   "store-1",
				Timestamp: "2025-01-10T10:15:00Z",
				Items:     []Item{},
			},
			wantErr: true,
		},
		{
			name: "Empty items",
			tx: Transaction{
				TransactionID: "tx-1",
				ShopperID:     "shopper-1",
				StoreID:       "store-1",
				Timestamp:     "2025-01-10T10:15:00Z",
				Items:         []Item{},
			},
			wantErr: true,
		},
		{
			name: "Negative quantity",
			tx: Transaction{
				TransactionID: "tx-1",
				ShopperID:     "shopper-1",
				StoreID:       "store-1",
				Timestamp:     "2025-01-10T10:15:00Z",
				Items: []Item{
					{SKU: "SKU-1", Name: "Item 1", Quantity: -1, UnitPrice: 10, Category: "grocery"},
				},
			},
			wantErr: true,
		},
		{
			name: "Negative price",
			tx: Transaction{
				TransactionID: "tx-1",
				ShopperID:     "shopper-1",
				StoreID:       "store-1",
				Timestamp:     "2025-01-10T10:15:00Z",
				Items: []Item{
					{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: -10, Category: "grocery"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransaction(&tt.tx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateTransaction(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	tx := Transaction{
		TransactionID: "tx-test-1",
		ShopperID:     "shopper-test-1",
		StoreID:       "store-1",
		Timestamp:     "2025-01-10T10:15:00Z",
		Items: []Item{
			{SKU: "SKU-MILK", Name: "Milk", Quantity: 2, UnitPrice: 5, Category: "grocery"},
			{SKU: "SKU-PLUSH", Name: "Promo Plush Toy", Quantity: 1, UnitPrice: 15, Category: "promo"},
		},
	}

	body, _ := json.Marshal(tx)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "tx-test-1", response["transaction_id"])
	assert.Equal(t, float64(3), response["stickers_earned"]) // $25 = 2 base, +1 promo = 3 total
}

func TestDuplicateTransactionIdempotency(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	tx := Transaction{
		TransactionID: "tx-duplicate",
		ShopperID:     "shopper-1",
		StoreID:       "store-1",
		Timestamp:     "2025-01-10T10:15:00Z",
		Items: []Item{
			{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 10, Category: "grocery"},
		},
	}

	body, _ := json.Marshal(tx)

	// First submission
	req1, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusCreated, w1.Code)

	var response1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &response1)
	stickers1 := response1["stickers_earned"]

	// Duplicate submission
	req2, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &response2)
	stickers2 := response2["stickers_earned"]

	// Should return same result
	assert.Equal(t, stickers1, stickers2)
	assert.Contains(t, response2["message"].(string), "already processed")
}

func TestGetShopperStatus(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Create a transaction
	tx := Transaction{
		TransactionID: "tx-status-1",
		ShopperID:     "shopper-status",
		StoreID:       "store-1",
		Timestamp:     "2025-01-10T10:15:00Z",
		Items: []Item{
			{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 20, Category: "grocery"},
		},
	}

	body, _ := json.Marshal(tx)
	req, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Get shopper status
	req2, _ := http.NewRequest("GET", "/shoppers/shopper-status", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var status ShopperStatus
	json.Unmarshal(w2.Body.Bytes(), &status)
	assert.Equal(t, "shopper-status", status.ShopperID)
	assert.Equal(t, 2, status.CurrentBalance) // $20 = 2 stickers
	assert.Equal(t, 1, len(status.Transactions))
}

func TestRedemption(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	// Create transactions to earn stickers (need at least 10 for a Mug)
	tx1 := Transaction{
		TransactionID: "tx-red-1",
		ShopperID:     "shopper-red",
		StoreID:       "store-1",
		Timestamp:     "2025-01-10T10:15:00Z",
		Items: []Item{
			{SKU: "SKU-1", Name: "Item 1", Quantity: 1, UnitPrice: 100, Category: "grocery"},
		},
	}

	body1, _ := json.Marshal(tx1)
	req1, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	tx2 := Transaction{
		TransactionID: "tx-red-2",
		ShopperID:     "shopper-red",
		StoreID:       "store-1",
		Timestamp:     "2025-01-10T10:16:00Z",
		Items: []Item{
			{SKU: "SKU-2", Name: "Item 2", Quantity: 1, UnitPrice: 100, Category: "grocery"},
		},
	}

	body2, _ := json.Marshal(tx2)
	req2, _ := http.NewRequest("POST", "/transactions", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Try to redeem
	redemption := RedemptionRequest{
		ShopperID:  "shopper-red",
		RewardName: "Mug",
	}

	body3, _ := json.Marshal(redemption)
	req3, _ := http.NewRequest("POST", "/redemptions", bytes.NewBuffer(body3))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusCreated, w3.Code)

	var response map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &response)
	assert.Equal(t, "Mug", response["reward_name"])
	assert.Equal(t, float64(10), response["stickers_cost"])
	assert.Equal(t, float64(0), response["new_balance"]) // 10 (5+5) - 10 = 0
}

func TestInsufficientStickersRedemption(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupRouter()

	redemption := RedemptionRequest{
		ShopperID:  "shopper-poor",
		RewardName: "Mug",
	}

	body, _ := json.Marshal(redemption)
	req, _ := http.NewRequest("POST", "/redemptions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response.Error, "Insufficient stickers")
}

