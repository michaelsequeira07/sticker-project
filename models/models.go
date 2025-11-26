package models

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

// CalculationBreakdown shows how stickers were calculated
type CalculationBreakdown struct {
	TotalAmount  float64 `json:"total_amount"`
	BaseStickers int     `json:"base_stickers"`
	PromoBonus   int     `json:"promo_bonus"`
	BeforeCap    int     `json:"before_cap"`
	AfterCap     int     `json:"after_cap"`
	Capped       bool    `json:"capped"`
}

// TransactionDebugInfo contains full debug information about a transaction
type TransactionDebugInfo struct {
	TransactionRecord
	Calculation    CalculationBreakdown `json:"calculation"`
	ShopperBalance int                  `json:"shopper_balance_after"`
	ProcessingTime string               `json:"processing_time"`
}

