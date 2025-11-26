package calculator

import "github.com/michaelsequeira07/sticker-project/models"

// CalculateStickers calculates stickers earned based on campaign rules
func CalculateStickers(items []models.Item) int {
	breakdown := CalculateStickersWithBreakdown(items)
	return breakdown.AfterCap
}

// CalculateStickersWithBreakdown calculates stickers and returns breakdown
func CalculateStickersWithBreakdown(items []models.Item) models.CalculationBreakdown {
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

	return models.CalculationBreakdown{
		TotalAmount:  totalAmount,
		BaseStickers: baseStickers,
		PromoBonus:   promoBonus,
		BeforeCap:    totalStickers,
		AfterCap:     afterCap,
		Capped:       capped,
	}
}

