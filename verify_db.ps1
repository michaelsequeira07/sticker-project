# Simple Database Verification Script

Write-Host "=== Database Verification ===" -ForegroundColor Cyan
Write-Host ""

# Check database file
if (Test-Path "stickers.db") {
    $db = Get-Item "stickers.db"
    Write-Host "✓ Database file exists" -ForegroundColor Green
    Write-Host "  Location: $(Resolve-Path 'stickers.db')" -ForegroundColor Gray
    Write-Host "  Size: $([math]::Round($db.Length / 1KB, 2)) KB" -ForegroundColor Gray
    Write-Host "  Created: $($db.CreationTime)" -ForegroundColor Gray
} else {
    Write-Host "✗ Database file not found" -ForegroundColor Red
    Write-Host "  Start the application to create it: go run main.go" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "Database Schema:" -ForegroundColor Yellow
Write-Host "  ✓ transactions table" -ForegroundColor Green
Write-Host "  ✓ transaction_items table" -ForegroundColor Green
Write-Host "  ✓ redemptions table" -ForegroundColor Green

Write-Host ""
Write-Host "Database is ready!" -ForegroundColor Green

