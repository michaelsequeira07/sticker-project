package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// InitDB initializes the database connection and creates tables
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
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
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

