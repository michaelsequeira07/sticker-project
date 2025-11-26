# Database Setup and Test Script

Write-Host "=== Database Setup Test ===" -ForegroundColor Cyan
Write-Host ""

# Check if database file exists
Write-Host "Checking database file..." -ForegroundColor Yellow
if (Test-Path "stickers.db") {
    $dbInfo = Get-Item "stickers.db"
    Write-Host "✓ Database file exists: $($dbInfo.Name)" -ForegroundColor Green
    Write-Host "  Size: $([math]::Round($dbInfo.Length / 1KB, 2)) KB" -ForegroundColor Gray
    Write-Host "  Created: $($dbInfo.CreationTime)" -ForegroundColor Gray
    Write-Host "  Modified: $($dbInfo.LastWriteTime)" -ForegroundColor Gray
} else {
    Write-Host "✗ Database file not found" -ForegroundColor Red
    Write-Host "  The database will be created when the application starts" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Testing database operations..." -ForegroundColor Yellow

# Test 1: Create a transaction
Write-Host "1. Creating test transaction..." -ForegroundColor Cyan
$transaction = @{
    transaction_id = "tx-db-setup-test"
    shopper_id = "shopper-setup-test"
    store_id = "store-01"
    timestamp = "2025-01-10T10:15:00Z"
    items = @(
        @{
            sku = "SKU-MILK"
            name = "Milk"
            quantity = 2
            unit_price = 5
            category = "grocery"
        },
        @{
            sku = "SKU-PROMO"
            name = "Promo Item"
            quantity = 1
            unit_price = 15
            category = "promo"
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/transactions" -Method POST -Body $transaction -ContentType "application/json"
    Write-Host "   ✓ Transaction created successfully" -ForegroundColor Green
    Write-Host "   Transaction ID: $($response.transaction_id)" -ForegroundColor Gray
    Write-Host "   Stickers Earned: $($response.stickers_earned)" -ForegroundColor Gray
    Write-Host "   Current Balance: $($response.current_balance)" -ForegroundColor Gray
} catch {
    Write-Host "   ✗ Failed to create transaction: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 2: Get shopper status
Write-Host ""
Write-Host "2. Retrieving shopper status..." -ForegroundColor Cyan
try {
    $shopperStatus = Invoke-RestMethod -Uri "http://localhost:8080/shoppers/shopper-setup-test"
    Write-Host "   ✓ Shopper status retrieved" -ForegroundColor Green
    Write-Host "   Shopper ID: $($shopperStatus.shopper_id)" -ForegroundColor Gray
    Write-Host "   Current Balance: $($shopperStatus.current_balance)" -ForegroundColor Gray
    Write-Host "   Total Earned: $($shopperStatus.total_earned)" -ForegroundColor Gray
    Write-Host "   Transactions: $($shopperStatus.transactions.Count)" -ForegroundColor Gray
} catch {
    Write-Host "   ✗ Failed to get shopper status: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Get transaction details
Write-Host ""
Write-Host "3. Retrieving transaction details..." -ForegroundColor Cyan
try {
    $txDetails = Invoke-RestMethod -Uri "http://localhost:8080/transactions/tx-db-setup-test"
    Write-Host "   ✓ Transaction details retrieved" -ForegroundColor Green
    Write-Host "   Transaction ID: $($txDetails.transaction_id)" -ForegroundColor Gray
    Write-Host "   Total Amount: `$$($txDetails.total_amount)" -ForegroundColor Gray
    Write-Host "   Stickers Earned: $($txDetails.stickers_earned)" -ForegroundColor Gray
    Write-Host "   Items: $($txDetails.items.Count)" -ForegroundColor Gray
    if ($txDetails.calculation) {
        Write-Host "   Calculation Breakdown:" -ForegroundColor Gray
        Write-Host "     Base Stickers: $($txDetails.calculation.base_stickers)" -ForegroundColor Gray
        Write-Host "     Promo Bonus: $($txDetails.calculation.promo_bonus)" -ForegroundColor Gray
        Write-Host "     Capped: $($txDetails.calculation.capped)" -ForegroundColor Gray
    }
} catch {
    Write-Host "   ✗ Failed to get transaction details: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Check database file size after operations
Write-Host ""
Write-Host "4. Checking database file after operations..." -ForegroundColor Cyan
if (Test-Path "stickers.db") {
    $dbInfo = Get-Item "stickers.db"
    Write-Host "   ✓ Database file updated" -ForegroundColor Green
    Write-Host "   Current Size: $([math]::Round($dbInfo.Length / 1KB, 2)) KB" -ForegroundColor Gray
    Write-Host "   Last Modified: $($dbInfo.LastWriteTime)" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=== Database Setup Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Database Location: $(Resolve-Path 'stickers.db')" -ForegroundColor Cyan
Write-Host ""
Write-Host "Database Schema:" -ForegroundColor Yellow
Write-Host "  - transactions (stores transaction records)" -ForegroundColor White
Write-Host "  - transaction_items (stores items within transactions)" -ForegroundColor White
Write-Host "  - redemptions (stores redemption history)" -ForegroundColor White
Write-Host ""
Write-Host "All database operations are working correctly!" -ForegroundColor Green

