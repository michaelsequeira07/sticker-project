package validation

import (
	"fmt"

	"github.com/michaelsequeira07/sticker-project/models"
)

// ValidateTransaction validates transaction input
func ValidateTransaction(tx *models.Transaction) error {
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

// ValidateRedemptionRequest validates redemption request input
func ValidateRedemptionRequest(req *models.RedemptionRequest) error {
	if req.ShopperID == "" {
		return fmt.Errorf("missing required field: shopper_id")
	}
	if req.RewardName == "" {
		return fmt.Errorf("missing required field: reward_name")
	}
	return nil
}

